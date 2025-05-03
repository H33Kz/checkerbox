package util

import "fmt"

type Queue[T any] struct {
	elements []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		elements: []T{},
	}
}

func (q *Queue[T]) Enqueue(element T) {
	q.elements = append(q.elements, element)
}

func (q *Queue[T]) Dequeue() T {
	element := q.elements[0]
	q.elements = q.elements[1:]
	return element
}

func (q *Queue[T]) PrintElements() {
	fmt.Println(q.elements)
}

func (q *Queue[T]) Len() int {
	return len(q.elements)
}

func (q *Queue[T]) Flush() {
	q.elements = []T{}
}
