// Package implements the singly-linked list.
//
// Structure is not thread safe.
//
// Reference: https://en.wikipedia.org/wiki/List_%28abstract_data_type%29
package util

import (
	"fmt"
	"strings"
)

// List holds the elements, where each element points to the next element
type List[T any] struct {
	first *element[T]
	last  *element[T]
	size  int
}

type element[T any] struct {
	value T
	next  *element[T]
}

// New instantiates a new list and adds the passed values, if any, to the list
func New[T any](values ...T) *List[T] {
	list := &List[T]{}
	if len(values) > 0 {
		list.Add(values...)
	}
	return list
}

// Add appends a value (one or more) at the end of the list (same as Append())
func (list *List[T]) Add(values ...T) {
	for _, value := range values {
		newElement := &element[T]{value: value}
		if list.size == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.last.next = newElement
			list.last = newElement
		}
		list.size++
	}
}

// Append appends a value (one or more) at the end of the list (same as Add())
func (list *List[T]) Append(values ...T) {
	list.Add(values...)
}

// Prepend prepends a values (or more)
func (list *List[T]) Prepend(values ...T) {
	// in reverse to keep passed order i.e. ["c","d"] -> Prepend(["a","b"]) -> ["a","b","c",d"]
	for v := len(values) - 1; v >= 0; v-- {
		newElement := &element[T]{value: values[v], next: list.first}
		list.first = newElement
		if list.size == 0 {
			list.last = newElement
		}
		list.size++
	}
}

// Get returns the element at index.
// Second return parameter is true if index is within bounds of the array and array is not empty, otherwise false.
func (list *List[T]) Get(index int) (T, bool) {

	if !list.withinRange(index) {
		var empty T
		return empty, false
	}

	element := list.first
	for e := 0; e != index; e, element = e+1, element.next {
	}

	return element.value, true
}

func (list *List[T]) Pop() (T, bool) {
	var empty T
	if list.size == 0 {
		return empty, false
	}
	element := list.first
	list.Remove(0)
	return element.value, true
}

// Remove removes the element at the given index from the list.
func (list *List[T]) Remove(index int) {

	if !list.withinRange(index) {
		return
	}

	if list.size == 1 {
		list.Clear()
		return
	}

	var beforeElement *element[T]
	element := list.first
	for e := 0; e != index; e, element = e+1, element.next {
		beforeElement = element
	}

	if element == list.first {
		list.first = element.next
	}
	if element == list.last {
		list.last = beforeElement
	}
	if beforeElement != nil {
		beforeElement.next = element.next
	}

	element = nil

	list.size--
}

// Values returns all elements in the list.
func (list *List[T]) Values() []T {
	values := make([]T, list.size)
	for e, element := 0, list.first; element != nil; e, element = e+1, element.next {
		values[e] = element.value
	}
	return values
}

// Empty returns true if list does not contain any elements.
func (list *List[T]) Empty() bool {
	return list.size == 0
}

// Size returns number of elements within the list.
func (list *List[T]) Size() int {
	return list.size
}

// Clear removes all elements from the list.
func (list *List[T]) Clear() {
	list.size = 0
	list.first = nil
	list.last = nil
}

// Swap swaps values of two elements at the given indices.
func (list *List[T]) Swap(i, j int) {
	if list.withinRange(i) && list.withinRange(j) && i != j {
		var element1, element2 *element[T]
		for e, currentElement := 0, list.first; element1 == nil || element2 == nil; e, currentElement = e+1, currentElement.next {
			switch e {
			case i:
				element1 = currentElement
			case j:
				element2 = currentElement
			}
		}
		element1.value, element2.value = element2.value, element1.value
	}
}

// Insert inserts values at specified index position shifting the value at that position (if any) and any subsequent elements to the right.
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List[T]) Insert(index int, values ...T) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(values...)
		}
		return
	}

	list.size += len(values)

	var beforeElement *element[T]
	foundElement := list.first
	for e := 0; e != index; e, foundElement = e+1, foundElement.next {
		beforeElement = foundElement
	}

	if foundElement == list.first {
		oldNextElement := list.first
		for i, value := range values {
			newElement := &element[T]{value: value}
			if i == 0 {
				list.first = newElement
			} else {
				beforeElement.next = newElement
			}
			beforeElement = newElement
		}
		beforeElement.next = oldNextElement
	} else {
		oldNextElement := beforeElement.next
		for _, value := range values {
			newElement := &element[T]{value: value}
			beforeElement.next = newElement
			beforeElement = newElement
		}
		beforeElement.next = oldNextElement
	}
}

// Set value at specified index
// Does not do anything if position is negative or bigger than list's size
// Note: position equal to list's size is valid, i.e. append.
func (list *List[T]) Set(index int, value T) {

	if !list.withinRange(index) {
		// Append
		if index == list.size {
			list.Add(value)
		}
		return
	}

	foundElement := list.first
	for e := 0; e != index; {
		e, foundElement = e+1, foundElement.next
	}
	foundElement.value = value
}

// String returns a string representation of container
func (list *List[T]) String() string {
	str := "SinglyLinkedList\n"
	values := []string{}
	for element := list.first; element != nil; element = element.next {
		values = append(values, fmt.Sprintf("%v", element.value))
	}
	str += strings.Join(values, ", ")
	return str
}

// Check that the index is within bounds of the list
func (list *List[T]) withinRange(index int) bool {
	return index >= 0 && index < list.size
}
