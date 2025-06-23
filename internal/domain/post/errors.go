package post

import "errors"

// 帖子相关错误
var (
	ErrPostNotFound      = errors.New("post not found")
	ErrPostAlreadyExists = errors.New("post already exists")
	ErrInvalidPostType   = errors.New("invalid post type")
	ErrInvalidPostStatus = errors.New("invalid post status")
	ErrInvalidTags       = errors.New("invalid tags format")
	ErrCannotPublish     = errors.New("cannot publish post")
	ErrCannotClose       = errors.New("cannot close post")
	ErrCannotEdit        = errors.New("cannot edit post")
	ErrTitleTooLong      = errors.New("title too long")
	ErrContentTooLong    = errors.New("content too long")
	ErrInvalidTokenID    = errors.New("invalid token ID")
	ErrUnauthorizedEdit  = errors.New("unauthorized to edit post")
)
