package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Prev  *ListItem
	Next  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	li := &ListItem{Value: v, Prev: nil, Next: l.front}
	if l.front != nil {
		l.front.Prev = li
	}
	if l.len == 0 {
		l.back = li
	}
	l.len++
	l.front = li
	return l.front
}

func (l *list) PushBack(v interface{}) *ListItem {
	li := &ListItem{Value: v, Prev: l.back, Next: nil}
	if l.back != nil {
		l.back.Next = li
	}
	if l.len == 0 {
		l.front = li
	}
	l.len++
	l.back = li
	return l.back
}

func (l *list) Remove(i *ListItem) {
	if (i != nil) && (l.len > 0) {
		switch {
		case l.len == 1:
			l.front = nil
			l.back = nil
		case i == l.front:
			i.Next.Prev = nil
			l.front = i.Next
		case i == l.back:
			i.Prev.Next = nil
			l.back = i.Prev
		default:
			i.Prev.Next = i.Next
			i.Next.Prev = i.Prev
		}
		l.len--
		i = nil
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if (i != nil) && (l.len > 1) && (i != l.front) {
		if i == l.back {
			l.back = i.Prev
			i.Prev.Next = nil
		} else {
			i.Prev.Next = i.Next
			i.Next.Prev = i.Prev
		}
		l.front.Prev = i
		i.Prev = nil
		i.Next = l.front
		l.front = i
	}
}

func NewList() List {
	return new(list)
}
