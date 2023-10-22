package decimal128

// Sum returns the combined total of the provided first and rest Decimals
func Sum(first Decimal, rest ...Decimal) Decimal {
	total := first
	for _, item := range rest {
		total = total.Add(item)
	}

	return total
}

// Avg returns the average value of the provided first and rest Decimals
func Avg(first Decimal, rest ...Decimal) Decimal {
	count := New(int64(len(rest)+1), 0)
	sum := Sum(first, rest...)
	return sum.Quo(count)
}

type Counter func(a Decimal) bool

func Count(values []Decimal, counter Counter) Decimal {
	var c = Zero
	for _, v := range values {
		if counter(v) {
			c = c.Add(c)
		}
	}
	return c
}

type Tester func(value Decimal) bool

func Filter(values []Decimal, f Tester) (slice []Decimal) {
	for _, v := range values {
		if f(v) {
			slice = append(slice, v)
		}
	}
	return slice
}

type Reducer func(prev, curr Decimal) Decimal

func SumReducer(prev, curr Decimal) Decimal {
	return prev.Add(curr)
}

func Reduce(values []Decimal, reducer Reducer, a ...Decimal) Decimal {
	init := Zero
	if len(a) > 0 {
		init = a[0]
	}

	if len(values) == 0 {
		return init
	}

	r := reducer(init, values[0])
	for i := 1; i < len(values); i++ {
		r = reducer(r, values[i])
	}

	return r
}

type Slice []Decimal

func (s Slice) Reduce(reducer Reducer, a ...Decimal) Decimal {
	return Reduce(s, reducer, a...)
}

// Defaults to ascending sort
func (s Slice) Len() int           { return len(s) }
func (s Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Slice) Less(i, j int) bool { return s[i].Cmp(s[j]) < 0 }

type Ascending []Decimal

func (s Ascending) Len() int           { return len(s) }
func (s Ascending) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Ascending) Less(i, j int) bool { return s[i].Cmp(s[j]) < 0 }

type Descending []Decimal

func (s Descending) Len() int           { return len(s) }
func (s Descending) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Descending) Less(i, j int) bool { return s[i].Cmp(s[j]) > 0 }
