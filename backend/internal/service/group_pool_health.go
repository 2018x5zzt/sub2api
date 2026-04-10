package service

import "math"

type GroupPoolHealth struct {
	AvailableAccountCount   int64
	RateLimitedAccountCount int64
	HealthPercent           int
	HealthState             string
}

func ComputeGroupPoolHealth(group *Group) GroupPoolHealth {
	if group == nil {
		return GroupPoolHealth{
			HealthState: "down",
		}
	}

	active := group.ActiveAccountCount
	if active < 0 {
		active = 0
	}

	limited := group.RateLimitedAccountCount
	if limited < 0 {
		limited = 0
	}

	available := active - limited
	if available < 0 {
		available = 0
	}

	percent := 0
	if denominator := available + limited; denominator > 0 {
		percent = int(math.Round(float64(available) * 100 / float64(denominator)))
	} else if available > 0 {
		percent = 100
	}

	state := "healthy"
	switch {
	case available <= 0:
		state = "down"
	case limited > 0:
		state = "degraded"
	}

	return GroupPoolHealth{
		AvailableAccountCount:   available,
		RateLimitedAccountCount: limited,
		HealthPercent:           percent,
		HealthState:             state,
	}
}
