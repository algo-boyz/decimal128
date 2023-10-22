package decimal128

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSum(t *testing.T) {
	vals := make([]Decimal, 10)
	var i = int64(0)

	for key := range vals {
		vals[key] = New(i, 0)
		i++
	}

	sum := Sum(vals[0], vals[1:]...)
	if !sum.Equal(New(45, 0)) {
		t.Errorf("Failed to calculate sum, expected %s got %s", New(45, 0), sum)
	}
}

func TestAvg(t *testing.T) {
	vals := make([]Decimal, 10)
	var i = int64(0)

	for key := range vals {
		vals[key] = New(i, 0)
		i++
	}

	avg := Avg(vals[0], vals[1:]...)
	if !avg.Equal(FromFloat64(4.5)) {
		t.Errorf("Failed to calculate average, expected %s got %s", FromFloat64(4.5).String(), avg.String())
	}
}

func TestReduce(t *testing.T) {
	type args struct {
		values  []Decimal
		init    Decimal
		reducer Reducer
	}
	tests := []struct {
		name string
		args args
		want Decimal
	}{
		{
			name: "simple",
			args: args{
				values: []Decimal{FromInt64(1), FromInt64(2), FromInt64(3)},
				init:   FromInt64(0),
				reducer: func(prev, curr Decimal) Decimal {
					return prev.Add(curr)
				},
			},
			want: FromInt64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Reduce(tt.args.values, tt.args.reducer, tt.args.init), "Reduce(%v, %v, %v)", tt.args.values, tt.args.init, tt.args.reducer)
		})
	}
}

func TestSortInterface(t *testing.T) {
	slice := Slice{
		FromInt64(7),
		FromInt64(3),
		FromInt64(1),
		FromInt64(2),
		FromInt64(5),
	}
	sort.Sort(slice)
	assert.Equal(t, "1", slice[0].String())
	assert.Equal(t, "2", slice[1].String())
	assert.Equal(t, "3", slice[2].String())
	assert.Equal(t, "5", slice[3].String())

	sort.Sort(Descending(slice))
	assert.Equal(t, "7", slice[0].String())
	assert.Equal(t, "5", slice[1].String())
	assert.Equal(t, "3", slice[2].String())
	assert.Equal(t, "2", slice[3].String())

	sort.Sort(Ascending(slice))
	assert.Equal(t, "1", slice[0].String())
	assert.Equal(t, "2", slice[1].String())
	assert.Equal(t, "3", slice[2].String())
	assert.Equal(t, "5", slice[3].String())
}
