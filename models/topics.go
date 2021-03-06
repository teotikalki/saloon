// Copyright 2018 The go-saloon Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/slices"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type Topic struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	Title       string      `json:"title" db:"title"`
	Content     string      `json:"content" db:"content"`
	AuthorID    uuid.UUID   `json:"author_id" db:"author_id"`
	CategoryID  uuid.UUID   `json:"category_id" db:"category_id"`
	Deleted     bool        `json:"deleted" db:"deleted"`
	Subscribers slices.UUID `json:"subscribers" db:"subscribers"`

	Author   *User     `json:"-" db:"-"`
	Category *Category `json:"-" db:"-"`
	Replies  Replies   `json:"-" db:"-"`
}

type Topics []Topic

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *Topic) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Content, Name: "Content"},
	), nil
}

func (t Topic) Authors() Users {
	var set = make(map[uuid.UUID]User, 1+len(t.Replies))
	set[t.Author.ID] = *t.Author
	for _, reply := range t.Replies {
		_, dup := set[reply.AuthorID]
		if dup {
			continue
		}
		if reply.Author != nil {
			set[reply.AuthorID] = *reply.Author
		}
	}

	authors := make([]User, 0, len(set))
	for _, v := range set {
		authors = append(authors, v)
	}
	return Users(authors)
}

func (t Topic) LastUpdate() time.Time {
	last := func(a, b time.Time) time.Time {
		if a.UTC().After(b.UTC()) {
			return a.UTC()
		}
		return b.UTC()
	}

	v := last(t.CreatedAt, t.UpdatedAt)
	for _, reply := range t.Replies {
		v = last(v, reply.CreatedAt)
		v = last(v, reply.UpdatedAt)
	}
	return v
}

func (t Topic) Subscribed(id uuid.UUID) bool {
	for _, usr := range t.Subscribers {
		if usr == id {
			return true
		}
	}
	return false
}

func (t *Topic) AddSubscriber(id uuid.UUID) {
	set := make(map[uuid.UUID]struct{})
	set[id] = struct{}{}
	for _, sub := range t.Subscribers {
		set[sub] = struct{}{}
	}
	subs := make(slices.UUID, 0, len(set))
	for sub := range set {
		subs = append(subs, sub)
	}
	t.Subscribers = subs
}

func (t *Topic) RemoveSubscriber(id uuid.UUID) {
	set := make(map[uuid.UUID]struct{})
	for _, sub := range t.Subscribers {
		if sub != id {
			set[sub] = struct{}{}
		}
	}
	subs := make(slices.UUID, 0, len(set))
	for sub := range set {
		subs = append(subs, sub)
	}
	t.Subscribers = subs
}
