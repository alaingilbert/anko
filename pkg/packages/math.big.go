package packages

import (
	"math/big"
)

func init() {
	Packages.Insert("math/big", PackageMap{
		"Above":        big.Above,
		"AwayFromZero": big.AwayFromZero,
		"Below":        big.Below,
		"Exact":        big.Exact,
		"Jacobi":       big.Jacobi,
		"MaxBase":      big.MaxBase,
		"MaxExp":       big.MaxExp,
		// TODO: https://github.com/alaingilbert/anko/issues/49
		// "MaxPrec":       big.MaxPrec,
		"MinExp":        big.MinExp,
		"NewFloat":      big.NewFloat,
		"NewInt":        big.NewInt,
		"NewRat":        big.NewRat,
		"ParseFloat":    big.ParseFloat,
		"ToNearestAway": big.ToNearestAway,
		"ToNearestEven": big.ToNearestEven,
		"ToNegativeInf": big.ToNegativeInf,
		"ToPositiveInf": big.ToPositiveInf,
		"ToZero":        big.ToZero,
	})
}
