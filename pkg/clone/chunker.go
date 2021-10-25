package clone

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strings"
	"sync"
)

// Chunk is an chunk of rows closed to the left [start,end)
type Chunk struct {
	Table *Table

	// Seq is the sequence number of chunks for this table
	Seq int64

	// Start is the first id of the chunk inclusive
	Start []interface{}
	// End is the first id of the next chunk (i.e. the last id of this chunk exclusively)
	End []interface{} // exclusive

	// First chunk of a table
	First bool
	// Last chunk of a table
	Last bool

	// Size is the expected number of rows in the chunk
	Size int
}

func (c *Chunk) String() string {
	return fmt.Sprintf("%s[%v-%v]", c.Table.Name, c.Start, c.End)
}

func (c *Chunk) ContainsRow(row []interface{}) bool {
	id := c.Table.PkOfRow(row)
	return c.ContainsPK(id)
}

func (c *Chunk) ContainsPK(id int64) bool {
	// TODO This is very odd because we support chunking by multiple columns but we don't support multiple primary keys. I
	//  think the path forward is to simply support multiple primary keys everywhere. That would require us to implement a
	//  composite key ordering function. I haven't thought this through deeply but I think we can simply do a multi value
	//  compare in Golang? It would become problematic for varchar columns with non-standard collations (those damn
	//  germans!) but I think we might be able to force utf8mb4 binary collation in the chunk select clause? I think as
	//  long as the ordering function is the same in the database as it is in Golang it's all good?
	return genericCompare(id, c.Start[c.Table.IDColumnInChunkColumns]) >= 0 && genericCompare(id, c.End[c.Table.IDColumnInChunkColumns]) < 0
}

func genericCompare(a interface{}, b interface{}) int {
	// Different database drivers interpret SQL types differently (it seems)
	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)

	// If they do NOT have same type, we coerce the target type to the source type and then compare
	// We only support the combinations we've encountered in the wild here
	switch a := a.(type) {
	case int64:
		coerced, err := coerceInt64(b)
		if err != nil {
			panic(err)
		}
		if a == coerced {
			return 0
		} else if a < coerced {
			return -1
		} else {
			return 1
		}
	case uint64:
		coerced, err := coerceUint64(b)
		if err != nil {
			panic(err)
		}
		if a == coerced {
			return 0
		} else if a < coerced {
			return -1
		} else {
			return 1
		}
	case float64:
		coerced, err := coerceFloat64(b)
		if err != nil {
			panic(err)
		}
		if a == coerced {
			return 0
		} else if a < coerced {
			return -1
		} else {
			return 1
		}
	case string:
		coerced, err := coerceString(b)
		if err != nil {
			panic(err)
		}
		if a == coerced {
			return 0
		} else if a < coerced {
			return -1
		} else {
			return 1
		}
	default:
		panic(fmt.Sprintf("type combination %v -> %v not supported yet: source=%v target=%v",
			aType, bType, a, b))
	}
}

func (c *Chunk) ContainsPKs(pk []interface{}) bool {
	// TODO when we support arbitrary primary keys this logic has to change
	if len(pk) != 1 {
		panic("currently only supported single integer pk")
	}
	i, err := coerceInt64(pk[0])
	if err != nil {
		panic(err)
	}
	return c.ContainsPK(i)
}

type PeekingIdStreamer interface {
	// Next returns next id and a boolean indicating if there is a next after this one
	Next(context.Context) ([]interface{}, bool, error)
	// Peek returns the id ahead of the current, Next above has to be called first
	Peek() []interface{}
}

type peekingIdStreamer struct {
	wrapped   IdStreamer
	peeked    []interface{}
	hasPeeked bool
}

func (p *peekingIdStreamer) Next(ctx context.Context) ([]interface{}, bool, error) {
	var err error
	if !p.hasPeeked {
		// first time round load the first entry
		p.peeked, err = p.wrapped.Next(ctx)
		if errors.Is(err, io.EOF) {
			return p.peeked, false, err
		} else {
			if err != nil {
				return p.peeked, false, errors.WithStack(err)
			}
		}
		p.hasPeeked = true
	}

	next := p.peeked
	hasNext := true

	p.peeked, err = p.wrapped.Next(ctx)
	if errors.Is(err, io.EOF) {
		hasNext = false
	} else {
		if err != nil {
			return next, hasNext, errors.WithStack(err)
		}
	}
	return next, hasNext, nil
}

func (p *peekingIdStreamer) Peek() []interface{} {
	return p.peeked
}

type IdStreamer interface {
	Next(context.Context) ([]interface{}, error)
}

type pagingStreamer struct {
	conn         DBReader
	first        bool
	currentPage  [][]interface{}
	currentIndex int
	pageSize     int
	retry        RetryOptions

	table        string
	chunkColumns []string
}

func newPagingStreamer(conn DBReader, table *Table, pageSize int, retry RetryOptions) *pagingStreamer {
	p := &pagingStreamer{
		conn:         conn,
		retry:        retry,
		first:        true,
		pageSize:     pageSize,
		currentPage:  nil,
		currentIndex: 0,
		chunkColumns: table.ChunkColumns,
		table:        table.Name,
	}

	return p
}

func (p *pagingStreamer) Next(ctx context.Context) ([]interface{}, error) {
	if p.currentIndex == len(p.currentPage) {
		var err error
		p.currentPage, err = p.loadPage(ctx)
		if errors.Is(err, io.EOF) {
			// Race condition, the table was emptied
			return nil, io.EOF
		}
		if err != nil {
			return nil, errors.WithStack(err)
		}
		p.currentIndex = 0
	}
	if len(p.currentPage) == 0 {
		return nil, io.EOF
	}
	next := p.currentPage[p.currentIndex]
	p.currentIndex++
	return next, nil
}

func (p *pagingStreamer) loadPage(ctx context.Context) ([][]interface{}, error) {
	var result [][]interface{}
	err := Retry(ctx, p.retry, func(ctx context.Context) error {
		var err error

		result = make([][]interface{}, 0, p.pageSize)
		var rows *sql.Rows
		if p.first {
			p.first = false
			chunkColumns := strings.Join(p.chunkColumns, ", ")
			stmt := fmt.Sprintf("select %s from %s order by %s limit %d",
				chunkColumns, p.table, chunkColumns, p.pageSize)
			rows, err = p.conn.QueryContext(ctx, stmt)
			if err != nil {
				return errors.Wrapf(err, "could not execute query: %v", stmt)
			}
			defer rows.Close()
		} else {
			result = nil
			if len(p.currentPage) == 0 {
				// Race condition, the table was emptied
				return backoff.Permanent(io.EOF)
			}
			lastItem := p.currentPage[len(p.currentPage)-1]
			comparison, params := expandRowConstructorComparison(p.chunkColumns, ">", lastItem)
			chunkColumns := strings.Join(p.chunkColumns, ", ")
			stmt := fmt.Sprintf("select %s from %s where %s order by %s limit %d",
				chunkColumns, p.table, comparison, chunkColumns, p.pageSize)
			rows, err = p.conn.QueryContext(ctx, stmt, params...)
			if err != nil {
				return errors.Wrapf(err, "could not execute query: %v", stmt)
			}
			defer rows.Close()
		}
		for rows.Next() {
			scanArgs := make([]interface{}, len(p.chunkColumns))
			for i := range scanArgs {
				// TODO support other types here
				scanArgs[i] = new(int64)
			}
			err := rows.Scan(scanArgs...)
			if err != nil {
				return errors.WithStack(err)
			}
			item := make([]interface{}, len(p.chunkColumns))
			for i := range scanArgs {
				item[i] = *scanArgs[i].(*int64)
			}
			result = append(result, item)
		}
		return err
	})

	return result, err
}

func streamIds(conn DBReader, table *Table, pageSize int, retry RetryOptions) PeekingIdStreamer {
	return &peekingIdStreamer{
		wrapped: newPagingStreamer(conn, table, pageSize, retry),
	}
}

func generateTableChunks(ctx context.Context, table *Table, source *sql.DB, retry RetryOptions) ([]Chunk, error) {
	var chunks []Chunk
	chunkCh := make(chan Chunk)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for c := range chunkCh {
			chunks = append(chunks, c)
		}
	}()
	err := generateTableChunksAsync(ctx, table, source, chunkCh, retry)
	close(chunkCh)
	if err != nil {
		return chunks, errors.WithStack(err)
	}
	wg.Wait()
	return chunks, nil
}

// generateTableChunksAsync generates chunks async on the current goroutine
func generateTableChunksAsync(ctx context.Context, table *Table, source *sql.DB, chunks chan Chunk, retry RetryOptions) error {
	chunkSize := table.Config.ChunkSize

	ids := streamIds(source, table, chunkSize, retry)

	var err error
	currentChunkSize := 0
	first := true
	var startId []interface{}
	seq := int64(0)
	var id []interface{}
	hasNext := true
	for hasNext {
		id, hasNext, err = ids.Next(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return errors.WithStack(err)
		}
		currentChunkSize++

		if startId == nil {
			startId = id
		}

		if currentChunkSize == chunkSize {
			chunksEnqueued.WithLabelValues(table.Name).Inc()
			nextId := ids.Peek()
			if !hasNext {
				nextId = nextChunkPosition(id)
			}
			select {
			case chunks <- Chunk{
				Table: table,
				Seq:   seq,
				Start: startId,
				End:   nextId,
				First: first,
				Last:  !hasNext,
				Size:  currentChunkSize,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
			seq++
			// Next id should be the next start id
			startId = nextId
			// We are no longer the first chunk
			first = false
			// We have no rows in the next chunk yet
			currentChunkSize = 0
		}
	}
	// Send any partial chunk we might have
	if currentChunkSize > 0 {
		chunksEnqueued.WithLabelValues(table.Name).Inc()
		select {
		case chunks <- Chunk{
			Table: table,
			Seq:   seq,
			Start: startId,
			// Make sure the End position is _after_ the final row by "adding one" to it
			End:   nextChunkPosition(id),
			First: first,
			Last:  true,
			Size:  currentChunkSize,
		}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func nextChunkPosition(pos []interface{}) []interface{} {
	result := make([]interface{}, len(pos))
	copy(result, pos)
	inc, err := increment(result[len(result)-1])
	if err != nil {
		panic(err)
	}
	result[len(result)-1] = inc
	return result
}

func increment(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case int64:
		return value + 1, nil
	default:
		return 0, errors.Errorf("can't (yet?) increment %v: %v", reflect.TypeOf(value), value)
	}
}
