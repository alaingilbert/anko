package packages

import (
	"strconv"
)

func init() {
	Packages.Insert("strconv", PackageMap{
		"FormatBool":  strconv.FormatBool,
		"FormatFloat": strconv.FormatFloat,
		"FormatInt":   strconv.FormatInt,
		"FormatUint":  strconv.FormatUint,
		"ParseBool":   strconv.ParseBool,
		"ParseFloat":  strconv.ParseFloat,
		"ParseInt":    strconv.ParseInt,
		"ParseUint":   strconv.ParseUint,
		"Atoi":        strconv.Atoi,
		"Itoa":        strconv.Itoa,
	})
}
