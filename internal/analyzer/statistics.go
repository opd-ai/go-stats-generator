package analyzer

import (
	"math"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ComputeLinearRegression performs least-squares linear regression on time series data.
// Returns regression coefficients (slope, intercept) and goodness of fit (R²).
func ComputeLinearRegression(series metrics.MetricTimeSeries) metrics.TrendStatistics {
	n := len(series.DataPoints)
	if n < 2 {
		return metrics.TrendStatistics{DataPoints: n}
	}

	x, y := convertToTimeSeries(series)
	slope, intercept := computeLeastSquares(x, y)
	rSquared := computeRSquared(x, y, slope, intercept, mean(y))
	correlation := computeCorrelation(rSquared, slope)

	return metrics.TrendStatistics{
		Slope:       slope,
		Intercept:   intercept,
		RSquared:    rSquared,
		DataPoints:  n,
		StartDate:   series.DataPoints[0].Timestamp.Format("2006-01-02"),
		EndDate:     series.DataPoints[n-1].Timestamp.Format("2006-01-02"),
		Correlation: correlation,
	}
}

func convertToTimeSeries(series metrics.MetricTimeSeries) (x, y []float64) {
	firstTime := series.DataPoints[0].Timestamp
	for _, pt := range series.DataPoints {
		days := pt.Timestamp.Sub(firstTime).Hours() / 24.0
		x = append(x, days)
		y = append(y, pt.Value)
	}
	return x, y
}

// computeLeastSquares calculates linear regression slope and intercept using least squares method.
func computeLeastSquares(x, y []float64) (slope, intercept float64) {
	meanX, meanY := mean(x), mean(y)

	var numerator, denominator float64
	for i := range x {
		numerator += (x[i] - meanX) * (y[i] - meanY)
		denominator += (x[i] - meanX) * (x[i] - meanX)
	}

	slope = 0.0
	if denominator != 0 {
		slope = numerator / denominator
	}
	intercept = meanY - slope*meanX
	return slope, intercept
}

// computeCorrelation derives Pearson correlation coefficient from R-squared and slope sign.
func computeCorrelation(rSquared, slope float64) float64 {
	correlation := 0.0
	if rSquared >= 0 {
		correlation = math.Sqrt(rSquared)
		if slope < 0 {
			correlation = -correlation
		}
	}
	return correlation
}

// computeRSquared calculates coefficient of determination (R²).
func computeRSquared(x, y []float64, slope, intercept, meanY float64) float64 {
	var ssRes, ssTot float64
	for i := range x {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	if ssTot == 0 {
		return 0
	}
	return 1 - (ssRes / ssTot)
}

// mean calculates arithmetic mean of a slice.
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// ComputePearsonCorrelation calculates the Pearson correlation coefficient between two datasets.
// Returns a MetricCorrelation containing the correlation, significance, and interpretive labels.
func ComputePearsonCorrelation(series1, series2 metrics.MetricTimeSeries) metrics.MetricCorrelation {
	x := extractValues(series1)
	y := extractValues(series2)
	n := len(x)
	if len(y) < n {
		n = len(y)
	}

	if n < 3 {
		return metrics.MetricCorrelation{
			Metric1:    series1.MetricName,
			Metric2:    series2.MetricName,
			DataPoints: n,
			Strength:   "insufficient_data",
			Direction:  "none",
		}
	}

	x, y = x[:n], y[:n]
	r := computePearsonR(x, y)
	pValue := computeCorrelationPValue(r, n)

	return metrics.MetricCorrelation{
		Metric1:     series1.MetricName,
		Metric2:     series2.MetricName,
		Correlation: r,
		PValue:      pValue,
		DataPoints:  n,
		Strength:    classifyCorrelationStrength(r),
		Direction:   classifyCorrelationDirection(r),
		Significant: pValue < 0.05,
	}
}

// extractValues gets the numeric values from a time series.
func extractValues(series metrics.MetricTimeSeries) []float64 {
	values := make([]float64, len(series.DataPoints))
	for i, pt := range series.DataPoints {
		values[i] = pt.Value
	}
	return values
}

// computePearsonR calculates the Pearson correlation coefficient.
func computePearsonR(x, y []float64) float64 {
	n := len(x)
	if n == 0 {
		return 0
	}

	meanX, meanY := mean(x), mean(y)
	var num, denX, denY float64

	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		num += dx * dy
		denX += dx * dx
		denY += dy * dy
	}

	if denX == 0 || denY == 0 {
		return 0
	}
	return num / math.Sqrt(denX*denY)
}

// computeCorrelationPValue approximates p-value using t-distribution.
func computeCorrelationPValue(r float64, n int) float64 {
	if n <= 2 || math.Abs(r) >= 1.0 {
		return 1.0
	}

	t := r * math.Sqrt(float64(n-2)/(1-r*r))
	df := float64(n - 2)

	// Approximate two-tailed p-value using Student's t-distribution
	return approximateTwoTailedPValue(math.Abs(t), df)
}

// approximateTwoTailedPValue computes an approximate p-value for t-statistic.
func approximateTwoTailedPValue(t, df float64) float64 {
	// Use a simple approximation for the t-distribution CDF
	x := df / (df + t*t)
	p := incompleteBetaApprox(df/2, 0.5, x)
	return p
}

// incompleteBetaApprox provides a simple approximation of the incomplete beta function.
func incompleteBetaApprox(a, b, x float64) float64 {
	// Simple series approximation for small x
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	// Use regularized incomplete beta approximation
	sum := 0.0
	term := 1.0
	for n := 0; n < 100; n++ {
		sum += term
		term *= (float64(n) + a) * x / (float64(n) + a + b)
		if math.Abs(term) < 1e-10 {
			break
		}
	}
	return sum * math.Pow(x, a) * math.Pow(1-x, b) / (a * beta(a, b))
}

// beta computes the beta function B(a,b) = Gamma(a)*Gamma(b)/Gamma(a+b).
func beta(a, b float64) float64 {
	return math.Gamma(a) * math.Gamma(b) / math.Gamma(a+b)
}

// classifyCorrelationStrength returns a label for the correlation strength.
func classifyCorrelationStrength(r float64) string {
	absR := math.Abs(r)
	if absR >= 0.7 {
		return "strong"
	}
	if absR >= 0.4 {
		return "moderate"
	}
	if absR >= 0.2 {
		return "weak"
	}
	return "none"
}

// classifyCorrelationDirection returns the direction label.
func classifyCorrelationDirection(r float64) string {
	if r > 0.1 {
		return "positive"
	}
	if r < -0.1 {
		return "negative"
	}
	return "none"
}

// ComputeCorrelationMatrix computes all pairwise correlations between metrics.
func ComputeCorrelationMatrix(seriesList []metrics.MetricTimeSeries) metrics.CorrelationMatrix {
	n := len(seriesList)
	if n < 2 {
		return metrics.CorrelationMatrix{}
	}

	metricNames := make([]string, n)
	for i, s := range seriesList {
		metricNames[i] = s.MetricName
	}

	correlations := make([]metrics.MetricCorrelation, 0, n*(n-1)/2)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			corr := ComputePearsonCorrelation(seriesList[i], seriesList[j])
			correlations = append(correlations, corr)
		}
	}

	dataPoints := 0
	startDate, endDate := "", ""
	if len(seriesList) > 0 && len(seriesList[0].DataPoints) > 0 {
		dataPoints = len(seriesList[0].DataPoints)
		startDate = seriesList[0].DataPoints[0].Timestamp.Format("2006-01-02")
		endDate = seriesList[0].DataPoints[dataPoints-1].Timestamp.Format("2006-01-02")
	}

	return metrics.CorrelationMatrix{
		Metrics:      metricNames,
		Correlations: correlations,
		DataPoints:   dataPoints,
		StartDate:    startDate,
		EndDate:      endDate,
	}
}
