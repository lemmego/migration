package migration

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Expr is a SQL fragment that is emitted verbatim rather than quoted as a
// literal. Use it for defaults that are expressions rather than values:
//
//	t.DateTime("created_at", 6).Default(Expr("CURRENT_TIMESTAMP"))
//	t.Int("attempts").Default(Expr("(1 + 1)"))
//
// Because the text is spliced into the statement unchanged, it must come from
// the migration itself and never from user input.
type Expr string

// currentTimestamp is the sentinel behind [CurrentTimestamp].
type currentTimestamp struct{}

// CurrentTimestamp is a default of "the moment the row is inserted". It is
// rendered as the expression each dialect expects, including the fractional
// second precision MySQL insists a DATETIME(n) default match:
//
//	t.DateTime("created_at", 6).Default(CurrentTimestamp)
//
// Spelling it as Expr("CURRENT_TIMESTAMP") instead is accepted by SQLite and
// PostgreSQL but rejected by MySQL with "Invalid default value", because there
// the precision of the default has to equal the precision of the column.
var CurrentTimestamp any = currentTimestamp{}

// defaultTimeLayout is understood as a timestamp literal by SQLite, MySQL and
// PostgreSQL alike. Sub-second precision is included only when it is non-zero,
// so a whole-second default does not gain a spurious ".000000".
const (
	defaultTimeLayout      = "2006-01-02 15:04:05"
	defaultTimeLayoutMicro = "2006-01-02 15:04:05.000000"
)

// defaultLiteral renders v as the SQL literal for a DEFAULT clause on the
// given dialect. It returns an error for types with no literal form, so the
// caller can report the problem rather than splice in a Go fmt rendering.
func defaultLiteral(dialect string, v any) (string, error) {
	return defaultLiteralWithPrecision(dialect, v, 0)
}

// defaultLiteralWithPrecision is defaultLiteral for a column whose type
// carries a fractional-second precision, which only [CurrentTimestamp] needs.
func defaultLiteralWithPrecision(dialect string, v any, precision uint) (string, error) {
	switch value := v.(type) {
	case nil:
		return "NULL", nil
	case currentTimestamp:
		return currentTimestampExpr(dialect, precision), nil
	case Expr:
		return string(value), nil
	case string:
		return quoteSQLString(value), nil
	case bool:
		return booleanLiteral(dialect, value), nil
	case time.Time:
		return quoteSQLString(formatDefaultTime(value)), nil
	case []byte:
		return blobLiteral(dialect, value), nil
	case int:
		return strconv.FormatInt(int64(value), 10), nil
	case int8:
		return strconv.FormatInt(int64(value), 10), nil
	case int16:
		return strconv.FormatInt(int64(value), 10), nil
	case int32:
		return strconv.FormatInt(int64(value), 10), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case uint:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(value), 10), nil
	case uint64:
		return strconv.FormatUint(value, 10), nil
	case float32:
		return floatLiteral(float64(value), 32)
	case float64:
		return floatLiteral(value, 64)
	default:
		return "", fmt.Errorf(
			"cannot use %T as a column default; pass a string, number, bool, time.Time, "+
				"[]byte or nil, or wrap a SQL expression in migration.Expr",
			v,
		)
	}
}

// quoteSQLString wraps s in single quotes, doubling any it contains. Doubling
// is the quote escape every SQL dialect accepts, and unlike a backslash escape
// it is unaffected by MySQL's NO_BACKSLASH_ESCAPES mode.
func quoteSQLString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// booleanLiteral renders a bool for the dialect's boolean representation.
// PostgreSQL has a real boolean type and rejects 1/0 for it; SQLite and MySQL
// store booleans as integers, where 1/0 is accepted by every version.
func booleanLiteral(dialect string, value bool) string {
	if dialect == DriverPostgres {
		if value {
			return "TRUE"
		}
		return "FALSE"
	}
	if value {
		return "1"
	}
	return "0"
}

func formatDefaultTime(t time.Time) string {
	if t.Nanosecond() == 0 {
		return t.Format(defaultTimeLayout)
	}
	return t.Format(defaultTimeLayoutMicro)
}

// blobLiteral renders a byte slice as a binary literal. SQLite and MySQL share
// the X'..' form; PostgreSQL needs a bytea hex escape string.
func blobLiteral(dialect string, value []byte) string {
	encoded := hex.EncodeToString(value)
	if dialect == DriverPostgres {
		return "'\\x" + encoded + "'"
	}
	return "X'" + encoded + "'"
}

// floatLiteral renders a float without an exponent where possible, since not
// every dialect accepts exponent notation in a DEFAULT clause. NaN and the
// infinities have no portable literal form at all.
func floatLiteral(value float64, bitSize int) (string, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "", fmt.Errorf(
			"%v has no portable SQL literal form and cannot be a column default", value,
		)
	}
	return strconv.FormatFloat(value, 'f', -1, bitSize), nil
}

// currentTimestampExpr renders the insert-time clock for the dialect. MySQL
// requires the default's precision to match the column's, so a DATETIME(6)
// column needs CURRENT_TIMESTAMP(6); SQLite and PostgreSQL take the bare form
// and derive the precision from the column.
func currentTimestampExpr(dialect string, precision uint) string {
	if dialect == DriverMySQL && precision > 0 {
		return "CURRENT_TIMESTAMP(" + strconv.FormatUint(uint64(precision), 10) + ")"
	}
	return "CURRENT_TIMESTAMP"
}
