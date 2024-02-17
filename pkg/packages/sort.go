package packages

import (
	"sort"
)

// SortFuncsStruct provides functions to be used with Sort
type SortFuncsStruct struct {
	LenFunc  func() int
	LessFunc func(i, j int) bool
	SwapFunc func(i, j int)
}

func (s SortFuncsStruct) Len() int           { return s.LenFunc() }
func (s SortFuncsStruct) Less(i, j int) bool { return s.LessFunc(i, j) }
func (s SortFuncsStruct) Swap(i, j int)      { s.SwapFunc(i, j) }

func init() {
	Packages.Insert("sort", PackageMap{
		"Float64s":          sort.Float64s,
		"Float64sAreSorted": sort.Float64sAreSorted,
		"Ints":              sort.Ints,
		"IntsAreSorted":     sort.IntsAreSorted,
		"IsSorted":          sort.IsSorted,
		"Search":            sort.Search,
		"SearchFloat64s":    sort.SearchFloat64s,
		"SearchInts":        sort.SearchInts,
		"SearchStrings":     sort.SearchStrings,
		"Sort":              sort.Sort,
		"Stable":            sort.Stable,
		"Strings":           sort.Strings,
		"StringsAreSorted":  sort.StringsAreSorted,
		"Slice":             sort.Slice,
		"SliceIsSorted":     sort.SliceIsSorted,
		"SliceStable":       sort.SliceStable,
	})
	PackageTypes.Insert("sort", PackageMap{
		"Float64Slice":    sort.Float64Slice{},
		"IntSlice":        sort.IntSlice{},
		"StringSlice":     sort.StringSlice{},
		"SortFuncsStruct": &SortFuncsStruct{},
	})
}
