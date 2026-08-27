package config

import "testing"

func TestStepsThroughOrdinal(t *testing.T) {
	steps := TransformPipeline{
		{RollingTotal: &RollingTotalConfig{}},
		{DateNormalize: &DateNormalizeConfig{Column: "date"}},
	}

	if got := StepsThroughOrdinal(steps, -1); len(got) != 0 {
		t.Fatalf("expected no steps for extract-only, got %d", len(got))
	}
	if got := StepsThroughOrdinal(steps, 0); len(got) != 2 {
		t.Fatalf("expected all steps for ordinal 0, got %d", len(got))
	}
	if got := StepsThroughOrdinal(steps, 1); len(got) != 1 {
		t.Fatalf("expected one step for ordinal 1, got %d", len(got))
	}
}
