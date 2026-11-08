package table

import (
	"fmt"
	"reflect"
	"strconv"
)

// Ordered is the set of types a column can hold and sort by.
type Ordered interface {
	int | int8 | int32 | int16 | int64 | float32 | float64 | string
}

// SortingOrderKey is the direction of a sort.
type SortingOrderKey int

// The two sort directions.
const (
	SortingOrderAscending  SortingOrderKey = iota // smallest first
	SortingOrderDescending                        // largest first
)

// GetOrder returns the sorted column, -1 when the rows are unsorted, and the direction.
func (r *Table) GetOrder() (int, SortingOrderKey) { return r.orderedColumnIndex, r.orderedColumnPhase }

// OrderByAsc sorts the rows by the column at index, smallest first, comparing values of
// the column's type. An index out of range, or fewer than two rows, changes nothing.
func (r *Table) OrderByAsc(index int) *Table {
	// sanity check first, we won't return errors here, simply ignore if the user sends non-existing index
	if index < len(r.columnHeaders) && len(r.filteredRows) > 1 {
		r.orderedColumnPhase = SortingOrderAscending
		r.rows = sortRows(r.rows, index, r.orderedColumnPhase)
		r.setRowsUpdate()
		r.orderedColumnIndex = index
		r.setHeadersUpdate()
	}
	return r
}

// OrderByDesc sorts the rows by the column at index, largest first, comparing values of
// the column's type. An index out of range, or fewer than two rows, changes nothing.
func (r *Table) OrderByDesc(index int) *Table {
	// sanity check first, we won't return errors here, simply ignore if the user sends non existing index
	if index < len(r.columnHeaders) && len(r.filteredRows) > 1 {
		r.orderedColumnPhase = SortingOrderDescending
		r.rows = sortRows(r.rows, index, r.orderedColumnPhase)
		r.setRowsUpdate()
		r.orderedColumnIndex = index
		r.setHeadersUpdate()
	}
	return r
}

func sortRows(rows [][]any, index int, orderKey SortingOrderKey) [][]any {
	// sorted rows
	var sorted [][]any
	// list of column values used for ordering
	var orderingCol []any
	for _, rw := range rows {
		orderingCol = append(orderingCol, rw[index])
	}
	// get sorting index
	sortingIndex := sortIndexByOrderedColumn(orderingCol, orderKey)
	// update rows
	for _, i := range sortingIndex {
		sorted = append(sorted, rows[i])
	}
	return sorted
}

// isOrdered check if type is one of valid Ordered types
func isOrdered(e any) bool {
	switch e.(type) {
	case string, int, int8, int16, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

// getStringFromOrdered returns string from interface that was produced with one of Ordered types
func getStringFromOrdered(i any) string {
	switch i := i.(type) {
	case string:
		return i
	case int:
		return strconv.Itoa(i)
	case int8:
		return strconv.Itoa(int(i))
	case int16:
		return strconv.Itoa(int(i))
	case int32:
		return strconv.Itoa(int(i))
	case int64:
		return strconv.Itoa(int(i))
	case float32:
		// default precision of 24
		return strconv.FormatFloat(float64(i), 'G', 0, 32)
	case float64:
		// default precision of 24
		return strconv.FormatFloat(i, 'G', 0, 64)
	default:
		return ""
	}
}

// sortIndexByOrderedColumn casts to the one of Ordered type that is used on the column and sends to sorting
// returns sorted index of elements rather than elements themselves
func sortIndexByOrderedColumn(i []any, order SortingOrderKey) (sortedIndex []int) {
	// if len of slice is 0 return empty sort order
	if len(i) == 0 {
		return sortedIndex
	}

	switch i[0].(type) {
	case string:
		var s []string
		for _, el := range i {
			s = append(s, el.(string))
		}
		return sortIndex(s, order)
	case int:
		var s []int
		for _, el := range i {
			s = append(s, el.(int))
		}
		return sortIndex(s, order)
	case int8:
		var s []int8
		for _, el := range i {
			s = append(s, el.(int8))
		}
		return sortIndex(s, order)
	case int16:
		var s []int16
		for _, el := range i {
			s = append(s, el.(int16))
		}
		return sortIndex(s, order)
	case int32:
		var s []int32
		for _, el := range i {
			s = append(s, el.(int32))
		}
		return sortIndex(s, order)
	case int64:
		var s []int64
		for _, el := range i {
			s = append(s, el.(int64))
		}
		return sortIndex(s, order)
	case float32:
		var s []float32
		for _, el := range i {
			s = append(s, el.(float32))
		}
		return sortIndex(s, order)
	case float64:
		var s []float64
		for _, el := range i {
			s = append(s, el.(float64))
		}
		return sortIndex(s, order)

	default:
		panic(fmt.Sprintf("type %s not subtype of Ordered", reflect.TypeOf(i[0]).String()))
	}
}

// sortIndex bubble-sorts slice in the given order and returns the original index of each
// element in its sorted position.
func sortIndex[T Ordered](slice []T, order SortingOrderKey) []int {
	// could do this in sortIndexByOrderedColumn where we cycle through the slice anyhow
	// tho I think this is cheap op and makes code a bit cleaner, worthy trade for now
	var index []int
	for i := 0; i < len(slice); i++ {
		index = append(index, i)
	}

	// bubble sort slice and update index in a process
	for i := len(slice); i > 0; i-- {
		for j := 1; j < i; j++ {
			if order == SortingOrderDescending && slice[j] > slice[j-1] {
				slice[j], slice[j-1] = slice[j-1], slice[j]
				index[j], index[j-1] = index[j-1], index[j]
			} else if order == SortingOrderAscending && slice[j] < slice[j-1] {
				slice[j], slice[j-1] = slice[j-1], slice[j]
				index[j], index[j-1] = index[j-1], index[j]
			}
		}
	}
	return index
}
