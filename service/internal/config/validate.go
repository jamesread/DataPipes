package config

import "fmt"

func validateConfig(c *Config) []string {
	if c == nil {
		return nil
	}

	var errs []string
	dir := ConfigDirectory()

	for _, id := range c.ConnectionNames() {
		conn := c.Connections[id]
		if conn == nil {
			continue
		}
		if err := validateConnection(id, conn); err != nil {
			errs = append(errs, err.Error())
		}
	}

	for _, jobID := range c.JobNames() {
		job := c.Job(jobID)
		if job == nil {
			continue
		}
		if job.Extract != "" {
			if c.Connections == nil || c.Connections[job.Extract] == nil {
				errs = append(errs, fmt.Sprintf("job %q: unknown extract connection %q", jobID, job.Extract))
			}
		}
		if job.Load != "" && !IsImplicitConnection(job.Load) {
			if c.Connections == nil || c.Connections[job.Load] == nil {
				errs = append(errs, fmt.Sprintf("job %q: unknown load connection %q", jobID, job.Load))
			}
		}

		eff := c.EffectiveConfigForJob(jobID)
		if eff == nil {
			continue
		}
		for i, step := range eff.Transform {
			if err := validateTransformStep(jobID, i+1, step, dir); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}

	return errs
}

func validateTransformStep(jobID string, stepNum int, step TransformStep, dir string) error {
	switch step.Kind() {
	case "":
		return fmt.Errorf("job %q transform step %d: unknown or empty transformation", jobID, stepNum)
	case "add_category":
		if _, err := step.AddCategory.Resolve(dir); err != nil {
			return fmt.Errorf("job %q transform step %d add_category: %v", jobID, stepNum, err)
		}
	case "date_normalize":
		if !step.DateNormalize.Configured() {
			return fmt.Errorf("job %q transform step %d date_normalize: column or columns required", jobID, stepNum)
		}
	case "date_to_incremental":
		if !step.DateToIncremental.Configured() {
			return fmt.Errorf("job %q transform step %d date_to_incremental: column required", jobID, stepNum)
		}
	case "append_hash":
		if !step.AppendHash.Configured() {
			return fmt.Errorf("job %q transform step %d append_hash: column required", jobID, stepNum)
		}
	}
	return nil
}

func validateConnection(id string, conn *Connection) error {
	switch conn.Type {
	case ConnectionTypeFireflyIII:
		if err := conn.ValidateFireflyIII(); err != nil {
			return fmt.Errorf("connection %q: %v", id, err)
		}
	}
	return nil
}
