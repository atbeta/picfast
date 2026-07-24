package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func PgInt8(v int64) pgtype.Int8 {
	return pgtype.Int8{Int64: v, Valid: true}
}

func PgInt8Ptr(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func PgInt8Val(p pgtype.Int8) int64 {
	if p.Valid {
		return p.Int64
	}
	return 0
}

func PgInt8PtrVal(p pgtype.Int8) *int64 {
	if p.Valid {
		return &p.Int64
	}
	return nil
}

func PgInt8PtrValI(v *int16) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: int64(*v), Valid: true}
}

func PgTimeWithZonePtr(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *v, Valid: true}
}

func PgTextNonEmpty(v string) pgtype.Text {
	if v == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: v, Valid: true}
}

func PgTextPtrVal(p pgtype.Text) *string {
	if p.Valid {
		return &p.String
	}
	return nil
}
