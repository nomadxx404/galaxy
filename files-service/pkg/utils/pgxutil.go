package pgxutil

import "github.com/jackc/pgx/v5/pgtype"

func TextNotValid(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}

	if *s == "" || *s == "null" || *s == "undefined" {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *s, Valid: true}
}

func TextValid(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func ToInt(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func ToInt64(i *int64) pgtype.Int8 {
	if i == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *i, Valid: true}
}

func ToBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func Pointer[T any](v T) *T {
	return &v
}
