package processor

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMovingAverageFilter_Comprehensive(t *testing.T) {
	f := NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()})
	assert.NotNil(t, f)
	assert.Equal(t, "ma", f.Name())

	r := f.Filter(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)
	assert.False(t, r.Filtered)

	r = f.Filter(20.0)
	assert.InDelta(t, 15.0, r.Value, 0.001)

	r = f.Filter(30.0)
	assert.InDelta(t, 20.0, r.Value, 0.001)
	assert.True(t, r.Filtered)

	f.Reset()
	r = f.Filter(5.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)
	assert.False(t, r.Filtered)
}

func TestMedianFilter_Comprehensive(t *testing.T) {
	f := NewMedianFilter(MedianFilterConfig{Name: "med", Window: 3, Logger: zap.NewNop()})
	assert.NotNil(t, f)
	assert.Equal(t, "med", f.Name())

	f.Filter(5.0)
	f.Filter(3.0)
	r := f.Filter(8.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)
	assert.True(t, r.Filtered)

	f.Reset()
	r = f.Filter(1.0)
	assert.InDelta(t, 1.0, r.Value, 0.001)
}

func TestKalmanFilter_Comprehensive(t *testing.T) {
	f := NewKalmanFilter(KalmanFilterConfig{Name: "kalman", ProcessNoise: 0.01, MeasureNoise: 0.1, Logger: zap.NewNop()})
	assert.NotNil(t, f)
	assert.Equal(t, "kalman", f.Name())

	r := f.Filter(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)
	assert.False(t, r.Filtered)

	r = f.Filter(12.0)
	assert.True(t, r.Filtered)
	assert.Greater(t, r.Value, 10.0)
	assert.Less(t, r.Value, 12.0)

	f.Reset()
	r = f.Filter(5.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)
}

func TestLimitFilter_Comprehensive(t *testing.T) {
	f := NewLimitFilter(LimitFilterConfig{Name: "limit", MaxChange: 5.0, Logger: zap.NewNop()})
	assert.NotNil(t, f)
	assert.Equal(t, "limit", f.Name())

	r := f.Filter(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)
	assert.False(t, r.Filtered)

	r = f.Filter(20.0)
	assert.InDelta(t, 15.0, r.Value, 0.001)
	assert.True(t, r.Filtered)

	r = f.Filter(5.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)
	assert.True(t, r.Filtered)

	f.Reset()
	r = f.Filter(50.0)
	assert.InDelta(t, 50.0, r.Value, 0.001)
}

func TestExponentialSmoothingFilter_Comprehensive(t *testing.T) {
	f := NewExponentialSmoothingFilter(ExponentialSmoothingFilterConfig{Name: "exp", Alpha: 0.3, Logger: zap.NewNop()})
	assert.NotNil(t, f)
	assert.Equal(t, "exp", f.Name())

	r := f.Filter(100.0)
	assert.InDelta(t, 100.0, r.Value, 0.001)
	assert.False(t, r.Filtered)

	r = f.Filter(110.0)
	assert.Greater(t, r.Value, 100.0)
	assert.Less(t, r.Value, 110.0)
	assert.True(t, r.Filtered)

	f.Reset()
	r = f.Filter(50.0)
	assert.InDelta(t, 50.0, r.Value, 0.001)
}

func TestFilterChain_Comprehensive(t *testing.T) {
	chain := NewFilterChain(FilterChainConfig{Name: "chain", Logger: zap.NewNop()})
	assert.NotNil(t, chain)
	assert.Equal(t, "chain", chain.Name())

	chain.AddFilter(NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()}))
	chain.AddFilter(NewLimitFilter(LimitFilterConfig{Name: "lim", MaxChange: 50.0, Logger: zap.NewNop()}))

	r := chain.Filter(50.0)
	assert.InDelta(t, 50.0, r.Value, 0.001)

	filters := chain.GetFilters()
	assert.Len(t, filters, 2)

	chain.RemoveFilter("ma")
	filters = chain.GetFilters()
	assert.Len(t, filters, 1)

	chain.Reset()
}

func TestPriorityFilter_Comprehensive(t *testing.T) {
	pf := NewPriorityFilter(PriorityFilterConfig{Name: "pf", Logger: zap.NewNop()})
	pf.AddFilter(NewLimitFilter(LimitFilterConfig{Name: "lim", MaxChange: 100.0, Logger: zap.NewNop()}), 1)
	pf.AddFilter(NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()}), 2)

	r := pf.Filter(50.0)
	assert.NotNil(t, r)

	pf.Reset()
	assert.Equal(t, "pf", pf.Name())
}

func TestBatchFilter_Comprehensive(t *testing.T) {
	bf := NewBatchFilter(NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()}), 2, zap.NewNop())
	values := []float64{10.0, 20.0, 30.0, 40.0, 50.0}
	results := bf.FilterBatch(values)
	assert.Len(t, results, 5)
}

func TestFilterFactory_Comprehensive(t *testing.T) {
	factory := NewFilterFactory(zap.NewNop())

	ma := factory.CreateMovingAverageFilter("ma", 5)
	assert.NotNil(t, ma)
	assert.Equal(t, "ma", ma.Name())

	med := factory.CreateMedianFilter("med", 5)
	assert.NotNil(t, med)

	kalman := factory.CreateKalmanFilter("kalman", 0.01, 0.1, 0.0)
	assert.NotNil(t, kalman)

	lim := factory.CreateLimitFilter("lim", 10.0)
	assert.NotNil(t, lim)

	exp := factory.CreateExponentialSmoothingFilter("exp", 0.3, 0.0)
	assert.NotNil(t, exp)

	chain := factory.CreateFilterChain("chain")
	assert.NotNil(t, chain)
}

func TestLinearScaler_Comprehensive(t *testing.T) {
	s := NewLinearScaler(LinearScalerConfig{Name: "linear", Slope: 2.0, Intercept: 1.0, Logger: zap.NewNop()})
	assert.NotNil(t, s)
	assert.Equal(t, "linear", s.Name())

	r := s.Scale(5.0)
	assert.InDelta(t, 11.0, r.Value, 0.001)
	assert.True(t, r.Scaled)

	r = s.Inverse(11.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)

	sZero := NewLinearScaler(LinearScalerConfig{Name: "zero", Slope: 0, Intercept: 0, Logger: zap.NewNop()})
	r = sZero.Inverse(5.0)
	assert.Equal(t, QualityBad|QualityReasonConfiguration, r.Quality)
}

func TestLinearScaler_RangeCheck(t *testing.T) {
	s := NewLinearScaler(LinearScalerConfig{
		Name: "range", Slope: 1.0, Intercept: 0.0,
		InputMin: 0, InputMax: 100, Logger: zap.NewNop(),
	})
	r := s.Scale(50.0)
	assert.True(t, r.Scaled)

	r = s.Scale(150.0)
	assert.True(t, r.Scaled)
}

func TestPolynomialScaler_Comprehensive(t *testing.T) {
	s := NewPolynomialScaler(PolynomialScalerConfig{Name: "poly", Coeffs: []float64{1.0, 2.0, 3.0}, Logger: zap.NewNop()})
	assert.NotNil(t, s)
	assert.Equal(t, "poly", s.Name())

	r := s.Scale(2.0)
	assert.InDelta(t, 17.0, r.Value, 0.001)
	assert.True(t, r.Scaled)

	r = s.Inverse(17.0)
	assert.True(t, r.Scaled)
}

func TestLookupTableScaler_Comprehensive(t *testing.T) {
	s := NewLookupTableScaler(LookupTableScalerConfig{
		Name: "lookup",
		Table: []LookupEntry{
			{Input: 0, Output: 0},
			{Input: 50, Output: 500},
			{Input: 100, Output: 1000},
		},
		Logger: zap.NewNop(),
	})
	assert.NotNil(t, s)
	assert.Equal(t, "lookup", s.Name())

	r := s.Scale(50.0)
	assert.InDelta(t, 500.0, r.Value, 0.001)

	r = s.Scale(25.0)
	assert.InDelta(t, 250.0, r.Value, 0.001)

	r = s.Scale(-10.0)
	assert.Equal(t, QualityUncertain|QualityReasonOutOfRange, r.Quality)

	r = s.Scale(150.0)
	assert.Equal(t, QualityUncertain|QualityReasonOutOfRange, r.Quality)

	sExtrap := NewLookupTableScaler(LookupTableScalerConfig{
		Name: "lookup_extrap",
		Table: []LookupEntry{
			{Input: 0, Output: 0},
			{Input: 100, Output: 1000},
		},
		Extrapolate: true,
		Logger:      zap.NewNop(),
	})
	r = sExtrap.Scale(-50.0)
	assert.Equal(t, QualityQuestionable|QualityReasonScaled, r.Quality)

	r = sExtrap.Scale(150.0)
	assert.Equal(t, QualityQuestionable|QualityReasonScaled, r.Quality)

	r = s.Inverse(250.0)
	assert.True(t, r.Scaled)

	emptyScaler := NewLookupTableScaler(LookupTableScalerConfig{Name: "empty", Table: []LookupEntry{}, Logger: zap.NewNop()})
	r = emptyScaler.Scale(5.0)
	assert.Equal(t, QualityBad|QualityReasonConfiguration, r.Quality)

	s.AddEntry(LookupEntry{Input: 75, Output: 750})
}

func TestEngineeringUnitScaler_Comprehensive(t *testing.T) {
	s := NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name: "c2f", FromUnit: "°C", ToUnit: "°F", Factor: 1.8, Offset: 32, Logger: zap.NewNop(),
	})
	assert.NotNil(t, s)
	assert.Equal(t, "c2f", s.Name())

	r := s.Scale(0.0)
	assert.InDelta(t, 32.0, r.Value, 0.001)
	assert.True(t, r.Scaled)

	r = s.Scale(100.0)
	assert.InDelta(t, 212.0, r.Value, 0.001)

	r = s.Inverse(212.0)
	assert.InDelta(t, 100.0, r.Value, 0.001)

	sZero := NewEngineeringUnitScaler(EngineeringUnitScalerConfig{
		Name: "zero", FromUnit: "X", ToUnit: "Y", Factor: 0, Offset: 0, Logger: zap.NewNop(),
	})
	r = sZero.Inverse(5.0)
	assert.Equal(t, QualityBad|QualityReasonConfiguration, r.Quality)
}

func TestScaleChain_Comprehensive(t *testing.T) {
	chain := NewScaleChain(ScaleChainConfig{Name: "sc", Logger: zap.NewNop()})
	chain.AddScaler(NewLinearScaler(LinearScalerConfig{Name: "l1", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()}))
	chain.AddScaler(NewLinearScaler(LinearScalerConfig{Name: "l2", Slope: 1.0, Intercept: 10.0, Logger: zap.NewNop()}))

	r := chain.Scale(5.0)
	assert.InDelta(t, 20.0, r.Value, 0.001)

	r = chain.Inverse(20.0)
	assert.True(t, r.Scaled)

	scalers := chain.GetScalers()
	assert.Len(t, scalers, 2)

	chain.RemoveScaler("l1")
	scalers = chain.GetScalers()
	assert.Len(t, scalers, 1)

	assert.Equal(t, "sc", chain.Name())
}

func TestBatchScaler_Comprehensive(t *testing.T) {
	bs := NewBatchScaler(NewLinearScaler(LinearScalerConfig{Name: "l", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()}), 2, zap.NewNop())
	values := []float64{1.0, 2.0, 3.0}
	results := bs.ScaleBatch(values)
	assert.Len(t, results, 3)
	assert.InDelta(t, 2.0, results[0].Value, 0.001)
	assert.InDelta(t, 4.0, results[1].Value, 0.001)
	assert.InDelta(t, 6.0, results[2].Value, 0.001)
}

func TestScalerFactory_Comprehensive(t *testing.T) {
	factory := NewScalerFactory(zap.NewNop())

	ls := factory.CreateLinearScaler("ls", 2.0, 1.0)
	assert.NotNil(t, ls)

	lsr := factory.CreateLinearScalerFromRange("lsr", 0, 100, 0, 1000)
	assert.NotNil(t, lsr)

	ps := factory.CreatePolynomialScaler("ps", []float64{1.0, 2.0})
	assert.NotNil(t, ps)

	lts := factory.CreateLookupTableScaler("lts", []LookupEntry{{Input: 0, Output: 0}}, false)
	assert.NotNil(t, lts)

	eus := factory.CreateEngineeringUnitScaler("eus", "C", "F", 1.8, 32)
	assert.NotNil(t, eus)

	sc := factory.CreateScaleChain("sc")
	assert.NotNil(t, sc)
}

func TestPredefinedConverters(t *testing.T) {
	c2f := CelsiusToFahrenheit(zap.NewNop())
	r := c2f.Scale(100.0)
	assert.InDelta(t, 212.0, r.Value, 0.001)

	f2c := FahrenheitToCelsius(zap.NewNop())
	r = f2c.Scale(212.0)
	assert.InDelta(t, 100.0, r.Value, 0.001)

	b2p := BarToPsi(zap.NewNop())
	r = b2p.Scale(1.0)
	assert.InDelta(t, 14.5038, r.Value, 0.001)

	p2b := PsiToBar(zap.NewNop())
	r = p2b.Scale(14.5038)
	assert.InDelta(t, 1.0, r.Value, 0.01)

	kw2mw := KWToMW(zap.NewNop())
	r = kw2mw.Scale(1000.0)
	assert.InDelta(t, 1.0, r.Value, 0.001)

	mw2kw := MWToKW(zap.NewNop())
	r = mw2kw.Scale(1.0)
	assert.InDelta(t, 1000.0, r.Value, 0.001)

	p2d := PercentToDecimal(zap.NewNop())
	r = p2d.Scale(75.0)
	assert.InDelta(t, 0.75, r.Value, 0.001)

	d2p := DecimalToPercent(zap.NewNop())
	r = d2p.Scale(0.75)
	assert.InDelta(t, 75.0, r.Value, 0.001)
}

func TestValidateScaleResult(t *testing.T) {
	err := ValidateScaleResult(ScaleResult{Value: 50.0}, 0, 100)
	assert.NoError(t, err)

	err = ValidateScaleResult(ScaleResult{Value: -1.0}, 0, 100)
	assert.Error(t, err)

	err = ValidateScaleResult(ScaleResult{Value: 101.0}, 0, 100)
	assert.Error(t, err)
}

func TestQualityCode_Comprehensive(t *testing.T) {
	assert.Equal(t, QualityCode(0x0000), QualityGood)
	assert.Equal(t, QualityCode(0x8000), QualityBad)
	assert.Equal(t, QualityCode(0x4000), QualityUncertain)
	assert.Equal(t, QualityCode(0x2000), QualityQuestionable)

	combined := QualityGood | QualityReasonFiltered
	assert.True(t, combined&QualityReasonFiltered != 0)
}

func TestQualityCodeString_All(t *testing.T) {
	s := QualityCodeString(QualityGood)
	assert.Contains(t, s, "GOOD")

	s = QualityCodeString(QualityBad | QualityReasonOutOfRange)
	assert.Contains(t, s, "BAD")
	assert.Contains(t, s, "OUT_OF_RANGE")

	s = QualityCodeString(QualityUncertain | QualityReasonFiltered)
	assert.Contains(t, s, "UNCERTAIN")
	assert.Contains(t, s, "FILTERED")
}

func TestQualityLevel_All(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())
	assert.Equal(t, QualityLevelGood, marker.GetLevel(QualityGood))
	assert.Equal(t, QualityLevelBad, marker.GetLevel(QualityBad))
	assert.Equal(t, QualityLevelPoor, marker.GetLevel(QualityUncertain))
	assert.Equal(t, QualityLevelFair, marker.GetLevel(QualityQuestionable))
}

func TestQualityMarker_Evaluate(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())

	info := marker.Evaluate(nil)
	assert.Equal(t, QualityBad|QualityReasonNull, info.Code)

	info = marker.Evaluate(math.NaN())
	assert.True(t, info.Code&QualityBad != 0)

	info = marker.Evaluate(0.0)
	assert.True(t, info.Code&QualityUncertain != 0)

	info = marker.Evaluate(42.0)
	assert.Equal(t, QualityGood, info.Code)

	info = marker.Evaluate("")
	assert.Equal(t, QualityBad|QualityReasonNull, info.Code)

	info = marker.Evaluate(struct{}{})
	assert.True(t, info.Code&QualityUncertain != 0)
}

func TestQualityMarker_Mark(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())
	info := marker.Mark(42.0, QualityBad|QualityReasonOutOfRange, "test reason")
	assert.Equal(t, QualityBad|QualityReasonOutOfRange, info.Code)
	assert.Contains(t, info.Reasons, "test reason")
}

func TestQualityMarker_Combine(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())

	combined := marker.Combine(QualityBad, QualityReasonOutOfRange)
	assert.True(t, combined&QualityBad != 0)

	combined = marker.Combine(QualityUncertain, QualityReasonFiltered)
	assert.True(t, combined&QualityUncertain != 0)

	combined = marker.Combine()
	assert.Equal(t, QualityGood, combined)
}

func TestQualityMarker_RecordQuality(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())
	marker.RecordQuality("point-1", QualityInfo{Code: QualityBad, Level: QualityLevelBad})
	marker.RecordQuality("point-1", QualityInfo{Code: QualityGood, Level: QualityLevelGood})

	history := marker.GetHistory("point-1")
	assert.Len(t, history, 2)

	history = marker.GetHistory("nonexistent")
	assert.Empty(t, history)
}

func TestQualityMarker_GetStatistics(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())
	marker.RecordQuality("point-1", QualityInfo{Code: QualityGood, Level: QualityLevelGood})
	marker.RecordQuality("point-1", QualityInfo{Code: QualityBad, Level: QualityLevelBad})

	stats := marker.GetStatistics("point-1")
	assert.Equal(t, 2, stats.TotalCount)
	assert.Equal(t, 1, stats.GoodCount)
	assert.Equal(t, 1, stats.BadCount)

	stats = marker.GetStatistics("nonexistent")
	assert.Equal(t, 0, stats.TotalCount)
}

func TestQualityChecker_Comprehensive(t *testing.T) {
	checker := NewQualityChecker(NewQualityMarker(zap.NewNop()), zap.NewNop())
	checker.AddRule(QualityCheckRule{
		Name:   "positive",
		Check:  func(v interface{}) bool { f, ok := toFloat64(v); return ok && f > 0 },
		Code:   QualityBad | QualityReasonOutOfRange,
		Reason: "value must be positive",
	})

	info := checker.Check(10.0)
	assert.Equal(t, QualityGood, info.Code)

	info = checker.Check(-5.0)
	assert.True(t, info.Code&QualityBad != 0)
}

func TestRangeValidator_Comprehensive(t *testing.T) {
	v := NewRangeValidator(RangeValidatorConfig{Name: "range", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()})
	assert.NotNil(t, v)
	assert.Equal(t, "range", v.Name())

	r := v.Validate(50.0)
	assert.True(t, r.Valid)

	r = v.Validate(-1.0)
	assert.False(t, r.Valid)

	r = v.Validate(101.0)
	assert.False(t, r.Valid)

	r = v.Validate(math.NaN())
	assert.False(t, r.Valid)

	r = v.Validate("not a number")
	assert.False(t, r.Valid)

	vExcl := NewRangeValidator(RangeValidatorConfig{Name: "excl", MinValue: 0, MaxValue: 100, Inclusive: false, Logger: zap.NewNop()})
	r = vExcl.Validate(0.0)
	assert.False(t, r.Valid)
	r = vExcl.Validate(50.0)
	assert.True(t, r.Valid)
}

func TestNullValidator_Comprehensive(t *testing.T) {
	v := NewNullValidator(NullValidatorConfig{Name: "null", AllowNull: false, AllowEmptyStr: false, AllowZero: false, Logger: zap.NewNop()})
	assert.NotNil(t, v)
	assert.Equal(t, "null", v.Name())

	r := v.Validate(nil)
	assert.False(t, r.Valid)

	r = v.Validate("")
	assert.False(t, r.Valid)

	r = v.Validate(0.0)
	assert.True(t, r.Valid)
	assert.NotEmpty(t, r.Warnings)

	r = v.Validate(42.0)
	assert.True(t, r.Valid)

	vAllow := NewNullValidator(NullValidatorConfig{Name: "null_allow", AllowNull: true, AllowEmptyStr: true, AllowZero: true, Logger: zap.NewNop()})
	r = vAllow.Validate(nil)
	assert.True(t, r.Valid)

	r = vAllow.Validate("")
	assert.True(t, r.Valid)
}

func TestTypeValidator_Comprehensive(t *testing.T) {
	v := NewTypeValidator(TypeValidatorConfig{Name: "type", ExpectType: reflect.TypeOf(float64(0)), Logger: zap.NewNop()})
	assert.NotNil(t, v)
	assert.Equal(t, "type", v.Name())

	r := v.Validate(10.0)
	assert.True(t, r.Valid)

	r = v.Validate(nil)
	assert.False(t, r.Valid)

	r = v.Validate(int32(10))
	assert.True(t, r.Valid)
	assert.NotEmpty(t, r.Warnings)

	r = v.Validate("string")
	assert.False(t, r.Valid)
}

func TestCustomValidator_Comprehensive(t *testing.T) {
	v := NewCustomValidator(CustomValidatorConfig{
		Name:       "custom",
		ValidateFn: func(v interface{}) (bool, string) { f, ok := toFloat64(v); if !ok || f <= 0 { return false, "must be positive" }; return true, "" },
		Logger:     zap.NewNop(),
	})
	assert.NotNil(t, v)
	assert.Equal(t, "custom", v.Name())

	r := v.Validate(10.0)
	assert.True(t, r.Valid)

	r = v.Validate(-1.0)
	assert.False(t, r.Valid)

	vNil := NewCustomValidator(CustomValidatorConfig{Name: "nil_fn", ValidateFn: nil, Logger: zap.NewNop()})
	r = vNil.Validate(10.0)
	assert.True(t, r.Valid)
	assert.NotEmpty(t, r.Warnings)
}

func TestValidatorChain_Comprehensive(t *testing.T) {
	chain := NewValidatorChain(ValidatorChainConfig{Name: "chain", StopOnFail: false, Logger: zap.NewNop()})
	chain.AddValidator(NewRangeValidator(RangeValidatorConfig{Name: "range", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()}))
	chain.AddValidator(NewNullValidator(NullValidatorConfig{Name: "null", Logger: zap.NewNop()}))

	r := chain.Validate(50.0)
	assert.True(t, r.Valid)

	r = chain.Validate(150.0)
	assert.False(t, r.Valid)

	validators := chain.GetValidators()
	assert.Len(t, validators, 2)

	chain.RemoveValidator("range")
	validators = chain.GetValidators()
	assert.Len(t, validators, 1)

	assert.Equal(t, "chain", chain.Name())
}

func TestBatchValidator_Comprehensive(t *testing.T) {
	bv := NewBatchValidator(NewRangeValidator(RangeValidatorConfig{Name: "range", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()}), 2, zap.NewNop())
	values := []interface{}{10.0, 50.0, 150.0, -5.0}
	results := bv.ValidateBatch(values)
	assert.Len(t, results, 4)
	assert.True(t, results[0].Valid)
	assert.True(t, results[1].Valid)
	assert.False(t, results[2].Valid)
	assert.False(t, results[3].Valid)
}

func TestValidatorFactory_Comprehensive(t *testing.T) {
	factory := NewValidatorFactory(zap.NewNop())

	rv := factory.CreateRangeValidator("rv", 0, 100, true)
	assert.NotNil(t, rv)

	nv := factory.CreateNullValidator("nv", false, false, false)
	assert.NotNil(t, nv)

	tv := factory.CreateTypeValidator("tv", reflect.TypeOf(float64(0)))
	assert.NotNil(t, tv)

	cv := factory.CreateCustomValidator("cv", func(v interface{}) (bool, string) { return true, "" })
	assert.NotNil(t, cv)

	vc := factory.CreateValidatorChain("vc", true)
	assert.NotNil(t, vc)
}

func TestBasicChangeDetector_Comprehensive(t *testing.T) {
	d := NewBasicChangeDetector(BasicChangeDetectorConfig{Name: "basic", Logger: zap.NewNop()})
	assert.NotNil(t, d)
	assert.Equal(t, "basic", d.Name())

	r := d.Detect(10.0)
	assert.False(t, r.Changed)

	r = d.Detect(20.0)
	assert.True(t, r.Changed)
	assert.Equal(t, ChangeTypeRising, r.Event.Type)

	r = d.Detect(5.0)
	assert.True(t, r.Changed)
	assert.Equal(t, ChangeTypeFalling, r.Event.Type)

	r = d.Detect(5.0)
	assert.False(t, r.Changed)

	d.Reset()
	r = d.Detect(10.0)
	assert.False(t, r.Changed)
}

func TestDeadbandChangeDetector_Comprehensive(t *testing.T) {
	d := NewDeadbandChangeDetector(DeadbandChangeDetectorConfig{Name: "db", Deadband: 5.0, Logger: zap.NewNop()})
	assert.NotNil(t, d)
	assert.Equal(t, "db", d.Name())

	r := d.Detect(10.0)
	assert.False(t, r.Changed)

	r = d.Detect(12.0)
	assert.False(t, r.Changed)

	r = d.Detect(20.0)
	assert.True(t, r.Changed)

	d.SetDeadband(10.0)
	r = d.Detect(25.0)
	assert.False(t, r.Changed)

	d.Reset()
	r = d.Detect(50.0)
	assert.False(t, r.Changed)
}

func TestDebouncer_Comprehensive(t *testing.T) {
	d := NewDebouncer(DebouncerConfig{Name: "deb", DebounceTime: 50 * time.Millisecond, Logger: zap.NewNop()})
	assert.NotNil(t, d)
	assert.Equal(t, "deb", d.Name())

	r := d.Process(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)

	r = d.Process(20.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)

	time.Sleep(60 * time.Millisecond)
	r = d.Process(20.0)
	assert.True(t, r.Changed)
	assert.InDelta(t, 20.0, r.Value, 0.001)

	assert.InDelta(t, 20.0, d.GetStableValue(), 0.001)

	d.SetDebounceTime(100 * time.Millisecond)
	d.Reset()
}

func TestBinaryChangeDetector_Comprehensive(t *testing.T) {
	d := NewBinaryChangeDetector(BinaryChangeDetectorConfig{Name: "bin", Logger: zap.NewNop()})
	assert.NotNil(t, d)
	assert.Equal(t, "bin", d.Name())

	r := d.Detect(true)
	assert.False(t, r.Changed)
	assert.InDelta(t, 1.0, r.Value, 0.001)

	r = d.Detect(false)
	assert.True(t, r.Changed)
	assert.Equal(t, ChangeTypeFalling, r.Event.Type)

	r = d.Detect(true)
	assert.True(t, r.Changed)
	assert.Equal(t, ChangeTypeRising, r.Event.Type)

	d.Reset()
}

func TestChangeDetectorWithDebounce_Comprehensive(t *testing.T) {
	inner := NewBasicChangeDetector(BasicChangeDetectorConfig{Name: "inner", Logger: zap.NewNop()})
	d := NewChangeDetectorWithDebounce(ChangeDetectorWithDebounceConfig{
		Name:         "debounce",
		Detector:     inner,
		DebounceTime: 50 * time.Millisecond,
		Logger:       zap.NewNop(),
	})
	assert.NotNil(t, d)
	assert.Equal(t, "debounce", d.Name())

	r := d.Detect(100.0)
	assert.False(t, r.Changed)

	d.Reset()
}

func TestChangeDetectorFactory_Comprehensive(t *testing.T) {
	factory := NewChangeDetectorFactory(zap.NewNop())

	bcd := factory.CreateBasicChangeDetector("basic")
	assert.NotNil(t, bcd)

	dcd := factory.CreateDeadbandChangeDetector("deadband", 1.0)
	assert.NotNil(t, dcd)

	deb := factory.CreateDebouncer("deb", 100*time.Millisecond)
	assert.NotNil(t, deb)

	soe := factory.CreateSOERecorder("soe", 100)
	assert.NotNil(t, soe)
	soe.Stop()

	bin := factory.CreateBinaryChangeDetector("bin")
	assert.NotNil(t, bin)
}

func TestSOERecorder_Comprehensive(t *testing.T) {
	recorder := NewSOERecorder(SOERecorderConfig{Name: "soe", MaxEvents: 100, Logger: zap.NewNop()})
	assert.NotNil(t, recorder)
	assert.Equal(t, "soe", recorder.Name())

	recorder.Record(ChangeEvent{Type: ChangeTypeRising, OldValue: 0, NewValue: 1, Timestamp: time.Now()}, "point-1", "Point 1", 1)
	recorder.Record(ChangeEvent{Type: ChangeTypeFalling, OldValue: 1, NewValue: 0, Timestamp: time.Now()}, "point-2", "Point 2", 2)

	time.Sleep(50 * time.Millisecond)

	events := recorder.GetEvents(10)
	assert.GreaterOrEqual(t, len(events), 2)

	eventsByPoint := recorder.GetEventsByPoint("point-1", 10)
	assert.GreaterOrEqual(t, len(eventsByPoint), 1)

	eventsByTime := recorder.GetEventsByTimeRange(time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	assert.GreaterOrEqual(t, len(eventsByTime), 2)

	if len(events) > 0 {
		err := recorder.Acknowledge(events[0].ID, "user1")
		assert.NoError(t, err)
	}

	recorder.Stop()
}

func TestChangeTypeString(t *testing.T) {
	assert.Equal(t, "none", changeTypeString(ChangeTypeNone))
	assert.Equal(t, "rising", changeTypeString(ChangeTypeRising))
	assert.Equal(t, "falling", changeTypeString(ChangeTypeFalling))
	assert.Equal(t, "toggle", changeTypeString(ChangeTypeToggle))
}

func TestGenerateEventID(t *testing.T) {
	id := generateEventID()
	assert.NotEmpty(t, id)
}

func TestAbs(t *testing.T) {
	assert.Equal(t, 5.0, abs(-5.0))
	assert.Equal(t, 5.0, abs(5.0))
	assert.Equal(t, 0.0, abs(0.0))
}

func TestProcessingContext_Comprehensive(t *testing.T) {
	ctx := NewProcessingContext(context.Background(), "point-1")
	assert.NotNil(t, ctx)
	assert.Equal(t, "point-1", ctx.PointID)

	ctx.SetMetadata("key1", "value1")
	val, ok := ctx.GetMetadata("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	_, ok = ctx.GetMetadata("nonexistent")
	assert.False(t, ok)
}

func TestPipeline_Comprehensive(t *testing.T) {
	pipeline := NewPipeline(PipelineConfig{Name: "test", Logger: zap.NewNop()})
	assert.NotNil(t, pipeline)

	pipeline.AddStage(NewValidationStage("validate", NewRangeValidator(RangeValidatorConfig{Name: "rv", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()}), zap.NewNop()))
	pipeline.AddStage(NewScaleStage("scale", NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()}), zap.NewNop()))

	data := &ProcessedData{Value: 10.0}
	result, err := pipeline.Process(context.Background(), data)
	require.NoError(t, err)
	assert.InDelta(t, 20.0, result.Value, 0.001)

	pipeline.RemoveStage("validate")
	stats := pipeline.GetStatistics()
	assert.Equal(t, int64(1), stats.TotalProcessed)
	assert.Equal(t, int64(1), stats.TotalSuccess)

	pipeline.ResetStatistics()
	stats = pipeline.GetStatistics()
	assert.Equal(t, int64(0), stats.TotalProcessed)
}

func TestPipeline_ProcessBatch(t *testing.T) {
	pipeline := NewPipeline(PipelineConfig{Name: "batch", Logger: zap.NewNop()})
	pipeline.AddStage(NewScaleStage("scale", NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()}), zap.NewNop()))

	dataList := []*ProcessedData{
		{Value: 10.0},
		{Value: 20.0},
		{Value: 30.0},
	}
	results, err := pipeline.ProcessBatch(context.Background(), dataList)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestPipeline_ParallelBatch(t *testing.T) {
	pipeline := NewPipeline(PipelineConfig{Name: "parallel", Parallel: true, Workers: 2, Logger: zap.NewNop()})
	pipeline.AddStage(NewScaleStage("scale", NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 3.0, Intercept: 0.0, Logger: zap.NewNop()}), zap.NewNop()))

	dataList := []*ProcessedData{
		{Value: 1.0},
		{Value: 2.0},
		{Value: 3.0},
		{Value: 4.0},
	}
	results, err := pipeline.ProcessBatch(context.Background(), dataList)
	require.NoError(t, err)
	assert.Len(t, results, 4)
}

func TestPipeline_FailedStage(t *testing.T) {
	pipeline := NewPipeline(PipelineConfig{Name: "fail", Logger: zap.NewNop()})
	pipeline.AddStage(NewValidationStage("validate", NewRangeValidator(RangeValidatorConfig{Name: "rv", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()}), zap.NewNop()))

	data := &ProcessedData{Value: 200.0}
	_, err := pipeline.Process(context.Background(), data)
	assert.Error(t, err)

	stats := pipeline.GetStatistics()
	assert.Equal(t, int64(1), stats.TotalFailed)
}

func TestPipelineBuilder_Comprehensive(t *testing.T) {
	builder := NewPipelineBuilder("built-pipeline", zap.NewNop())
	pipeline := builder.
		WithParallel(2).
		AddValidationStage("validate", NewRangeValidator(RangeValidatorConfig{Name: "rv", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()})).
		AddFilterStage("filter", NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()})).
		AddScaleStage("scale", NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()})).
		AddChangeDetectionStage("detect", NewBasicChangeDetector(BasicChangeDetectorConfig{Name: "bcd", Logger: zap.NewNop()})).
		AddQualityMarkStage("quality", NewQualityMarker(zap.NewNop())).
		AddCustomStage("custom", func(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
			data.Value = data.Value * 2
			return data, nil
		}).
		Build()

	assert.NotNil(t, pipeline)
}

func TestDataProcessor_Comprehensive(t *testing.T) {
	dp := NewDataProcessor(zap.NewNop())
	assert.NotNil(t, dp)

	pipeline := NewPipeline(PipelineConfig{Name: "p", Logger: zap.NewNop()})
	pipeline.AddStage(NewScaleStage("scale", NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()}), zap.NewNop()))
	dp.AddPipeline("test", pipeline)

	retrieved, ok := dp.GetPipeline("test")
	assert.True(t, ok)
	assert.NotNil(t, retrieved)

	_, ok = dp.GetPipeline("nonexistent")
	assert.False(t, ok)

	result, err := dp.Process(context.Background(), "test", &ProcessedData{Value: 5.0})
	require.NoError(t, err)
	assert.InDelta(t, 10.0, result.Value, 0.001)

	_, err = dp.Process(context.Background(), "nonexistent", &ProcessedData{Value: 5.0})
	assert.Error(t, err)

	allStats := dp.GetAllStatistics()
	assert.NotNil(t, allStats)

	dp.RemovePipeline("test")
	_, ok = dp.GetPipeline("test")
	assert.False(t, ok)
}

func TestPluginManager_Comprehensive(t *testing.T) {
	pm := NewPluginManager(zap.NewNop())
	assert.NotNil(t, pm)

	plugin := &testPlugin{name: "test-plugin", version: "1.0"}
	err := pm.RegisterPlugin(plugin)
	assert.NoError(t, err)

	err = pm.RegisterPlugin(plugin)
	assert.Error(t, err)

	p, ok := pm.GetPlugin("test-plugin")
	assert.True(t, ok)
	assert.Equal(t, "test-plugin", p.Name())

	stage, err := pm.CreateStageFromPlugin("test-plugin", map[string]interface{}{})
	require.NoError(t, err)
	assert.NotNil(t, stage)

	_, err = pm.CreateStageFromPlugin("nonexistent", nil)
	assert.Error(t, err)

	pm.UnregisterPlugin("test-plugin")
	_, ok = pm.GetPlugin("test-plugin")
	assert.False(t, ok)
}

type testPlugin struct {
	name    string
	version string
}

func (p *testPlugin) Name() string    { return p.name }
func (p *testPlugin) Version() string { return p.version }
func (p *testPlugin) CreateStage(config map[string]interface{}) (Stage, error) {
	return NewCustomStage("test-stage", func(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
		return data, nil
	}, zap.NewNop()), nil
}

func TestCustomStage_Comprehensive(t *testing.T) {
	stage := NewCustomStage("custom", func(ctx context.Context, data *ProcessedData) (*ProcessedData, error) {
		data.Value = data.Value * 3.0
		return data, nil
	}, zap.NewNop())

	input := &ProcessedData{Value: 10.0}
	result, err := stage.Process(context.Background(), input)
	require.NoError(t, err)
	assert.InDelta(t, 30.0, result.Value, 0.001)
	assert.Equal(t, "custom", stage.Name())
}

func TestMedianFilterOptimized_Comprehensive(t *testing.T) {
	f := NewMedianFilterOptimized("opt_med", 5, zap.NewNop())
	assert.NotNil(t, f)
	assert.Equal(t, "opt_med", f.Name())

	r := f.Filter(5.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)

	r = f.Filter(3.0)
	r = f.Filter(8.0)
	r = f.Filter(1.0)
	r = f.Filter(9.0)
	assert.Greater(t, r.Value, 0.0)

	f.Reset()
	r = f.Filter(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)
}

func TestQualityMarkStage_Comprehensive(t *testing.T) {
	marker := NewQualityMarker(zap.NewNop())
	stage := NewQualityMarkStage("qmark", marker, zap.NewNop())
	assert.NotNil(t, stage)
	assert.Equal(t, "qmark", stage.Name())

	data := &ProcessedData{Value: 42.0}
	result, err := stage.Process(context.Background(), data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestChangeDetectionStage_Comprehensive(t *testing.T) {
	detector := NewBasicChangeDetector(BasicChangeDetectorConfig{Name: "bcd", Logger: zap.NewNop()})
	stage := NewChangeDetectionStage("detect", detector, zap.NewNop())
	assert.NotNil(t, stage)
	assert.Equal(t, "detect", stage.Name())

	data := &ProcessedData{Value: 10.0}
	result, err := stage.Process(context.Background(), data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFilterStage_Comprehensive(t *testing.T) {
	filter := NewMovingAverageFilter(MovingAverageFilterConfig{Name: "ma", Window: 3, Logger: zap.NewNop()})
	stage := NewFilterStage("filter", filter, zap.NewNop())
	assert.NotNil(t, stage)
	assert.Equal(t, "filter", stage.Name())

	data := &ProcessedData{Value: 10.0}
	result, err := stage.Process(context.Background(), data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestScaleStage_Comprehensive(t *testing.T) {
	scaler := NewLinearScaler(LinearScalerConfig{Name: "ls", Slope: 2.0, Intercept: 0.0, Logger: zap.NewNop()})
	stage := NewScaleStage("scale", scaler, zap.NewNop())
	assert.NotNil(t, stage)
	assert.Equal(t, "scale", stage.Name())

	data := &ProcessedData{Value: 5.0}
	result, err := stage.Process(context.Background(), data)
	require.NoError(t, err)
	assert.InDelta(t, 10.0, result.Value, 0.001)
}

func TestValidationStage_Comprehensive(t *testing.T) {
	validator := NewRangeValidator(RangeValidatorConfig{Name: "rv", MinValue: 0, MaxValue: 100, Inclusive: true, Logger: zap.NewNop()})
	stage := NewValidationStage("validate", validator, zap.NewNop())
	assert.NotNil(t, stage)
	assert.Equal(t, "validate", stage.Name())

	data := &ProcessedData{Value: 50.0}
	result, err := stage.Process(context.Background(), data)
	require.NoError(t, err)
	assert.NotNil(t, result)

	data = &ProcessedData{Value: 200.0}
	_, err = stage.Process(context.Background(), data)
	assert.Error(t, err)
}

func TestDefaultFilterConfig(t *testing.T) {
	f := NewMovingAverageFilter(MovingAverageFilterConfig{Name: "default", Logger: zap.NewNop()})
	assert.NotNil(t, f)
	r := f.Filter(10.0)
	assert.InDelta(t, 10.0, r.Value, 0.001)

	kf := NewKalmanFilter(KalmanFilterConfig{Name: "default", Logger: zap.NewNop()})
	assert.NotNil(t, kf)

	ef := NewExponentialSmoothingFilter(ExponentialSmoothingFilterConfig{Name: "default", Logger: zap.NewNop()})
	assert.NotNil(t, ef)
}

func TestDefaultScalerConfig(t *testing.T) {
	s := NewLinearScaler(LinearScalerConfig{Name: "default", Slope: 1.0, Intercept: 0.0, Logger: zap.NewNop()})
	assert.NotNil(t, s)
	r := s.Scale(5.0)
	assert.InDelta(t, 5.0, r.Value, 0.001)
}
