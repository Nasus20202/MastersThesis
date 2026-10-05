package analysis

import (
	"path/filepath"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/jsonfile"
)

// ModelMean is a condition's macro score averaged over models, each model
// weighted equally. Difference pairs every model with the reference attempts
// of the same model: per scenario it averages the per-model differences over
// the models that ran the scenario in both, with a 95% bootstrap interval over
// scenarios.
type ModelMean struct {
	Models         []string `json:"models"`
	Macro          float64  `json:"macro"`
	ReferenceMacro *float64 `json:"reference_macro,omitempty"`
	// ReferenceMacros holds each model's reference macro score.
	ReferenceMacros map[string]float64 `json:"reference_macros,omitempty"`
	Difference      *Difference        `json:"macro_difference,omitempty"`
}

// ByModel splits attempts by the model that ran them, in model order.
func ByModel(attempts []RunAttempt) ([]string, map[string][]RunAttempt) {
	groups := map[string][]RunAttempt{}
	for _, attempt := range attempts {
		groups[attempt.Model] = append(groups[attempt.Model], attempt)
	}
	models := make([]string, 0, len(groups))
	for model := range groups {
		models = append(models, model)
	}
	slices.Sort(models)
	return models, groups
}

// MeanOverModels averages the condition's per-model macro scores and, when
// reference attempts exist, the reference's and the paired difference.
func MeanOverModels(attempts, reference []RunAttempt) ModelMean {
	models, groups := ByModel(attempts)
	_, referenceGroups := ByModel(reference)
	result := ModelMean{Models: models, ReferenceMacros: map[string]float64{}}
	var referenceTotal float64
	referenceModels := 0
	differences := map[string][]float64{}
	for _, model := range models {
		macro, _, _ := outcome(groups[model])
		result.Macro += macro
		if len(referenceGroups[model]) == 0 {
			continue
		}
		referenceMacro, _, _ := outcome(referenceGroups[model])
		result.ReferenceMacros[model] = referenceMacro
		referenceTotal += referenceMacro
		referenceModels++
		means, referenceMeans := scenarioMeans(groups[model]), scenarioMeans(referenceGroups[model])
		for scenario, score := range means {
			if referenceScore, ok := referenceMeans[scenario]; ok {
				differences[scenario] = append(differences[scenario], score-referenceScore)
			}
		}
	}
	if len(models) > 0 {
		result.Macro /= float64(len(models))
	}
	if referenceModels > 0 {
		mean := referenceTotal / float64(referenceModels)
		result.ReferenceMacro = &mean
	}
	values := make([]float64, 0, len(differences))
	for _, perModel := range differences {
		total := 0.0
		for _, value := range perModel {
			total += value
		}
		values = append(values, total/float64(len(perModel)))
	}
	result.Difference = pairedDifference(values)
	return result
}

// WriteModelMean stores model_mean.json next to the per-model summaries.
func WriteModelMean(dir string, mean ModelMean) error {
	return jsonfile.Write(filepath.Join(dir, "model_mean.json"), mean)
}
