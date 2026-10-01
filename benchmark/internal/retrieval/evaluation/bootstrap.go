package evaluation

import (
	"math/rand/v2"
	"slices"
)

// The paired bootstrap resamples probes with replacement and uses the same
// resamples for every comparison; the fixed seed makes reruns identical.
const (
	bootstrapResamples = 10000
	bootstrapSeed      = 1
)

// Difference is a configuration's rewrite hit@k minus the selected one's,
// with a 95% percentile bootstrap confidence interval over probes.
type Difference struct {
	Config string  `json:"config"`
	Mean   float64 `json:"mean"`
	Low    float64 `json:"ci_low"`
	High   float64 `json:"ci_high"`
}

func compareToSelected(selected ConfigScore, configs []ConfigScore) []Difference {
	var differences []Difference
	for _, config := range configs {
		if config.Name() == selected.Name() {
			continue
		}
		perProbe := make([]float64, len(config.Probes))
		for index := range perProbe {
			perProbe[index] = config.Probes[index].Rewrite.Hit - selected.Probes[index].Rewrite.Hit
		}
		low, high := bootstrapInterval(perProbe)
		differences = append(differences, Difference{Config: config.Name(), Mean: mean(perProbe), Low: low, High: high})
	}
	return differences
}

func bootstrapInterval(values []float64) (float64, float64) {
	random := rand.New(rand.NewPCG(bootstrapSeed, bootstrapSeed)) //nolint:gosec // Resampling must be seeded and reproducible.
	means := make([]float64, bootstrapResamples)
	sample := make([]float64, len(values))
	for resample := range means {
		for index := range sample {
			sample[index] = values[random.IntN(len(values))]
		}
		means[resample] = mean(sample)
	}
	slices.Sort(means)
	return means[bootstrapResamples*25/1000], means[bootstrapResamples*975/1000]
}

func mean(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}
