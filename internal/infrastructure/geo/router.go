package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Region represents a deployment region
type Region struct {
	ID       string
	Name     string
	Location string
	Endpoint string
	Priority int
}

// GeoRouter routes requests to appropriate regions based on tenant and latency
type GeoRouter struct {
	regions      []Region
	tenantMap    map[string]string // tenantID -> regionID
	latencyCache map[string]time.Duration
}

// NewGeoRouter creates a new geo router
func NewGeoRouter(regions []Region) *GeoRouter {
	return &GeoRouter{
		regions:      regions,
		tenantMap:    make(map[string]string),
		latencyCache: make(map[string]time.Duration),
	}
}

// GetRegion returns the appropriate region for a tenant
func (r *GeoRouter) GetRegion(tenantID uuid.UUID, requestLocation *Location) Region {
	// Check data residency requirements
	if regionID, exists := r.tenantMap[tenantID.String()]; exists {
		return r.getRegionByID(regionID)
	}

	// Route to lowest latency region
	return r.findLowestLatencyRegion(requestLocation)
}

// getRegionByID returns a region by ID
func (r *GeoRouter) getRegionByID(id string) Region {
	for _, region := range r.regions {
		if region.ID == id {
			return region
		}
	}
	// Return first region as default
	if len(r.regions) > 0 {
		return r.regions[0]
	}
	return Region{}
}

// findLowestLatencyRegion finds the region with lowest latency
func (r *GeoRouter) findLowestLatencyRegion(location *Location) Region {
	// In a real implementation, this would measure actual latency
	// For now, return the first region
	if len(r.regions) > 0 {
		return r.regions[0]
	}
	return Region{}
}

// SetTenantRegion sets the data residency for a tenant
func (r *GeoRouter) SetTenantRegion(tenantID uuid.UUID, regionID string) {
	r.tenantMap[tenantID.String()] = regionID
}

// Location represents a geographic location
type Location struct {
	Latitude  float64
	Longitude float64
	Country   string
}

// MultiRegionConfig holds configuration for multi-region deployment
type MultiRegionConfig struct {
	Regions []RegionConfig `yaml:"regions"`
}

// RegionConfig holds configuration for a single region
type RegionConfig struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Location string         `yaml:"location"`
	Services ServicesConfig `yaml:"services"`
}

// ServicesConfig holds service endpoints for a region
type ServicesConfig struct {
	API           string `yaml:"api"`
	NATS          string `yaml:"nats"`
	MongoDB       string `yaml:"mongodb"`
	Redis         string `yaml:"redis"`
	MinIO         string `yaml:"minio"`
	Elasticsearch string `yaml:"elasticsearch"`
}

// LoadBalancer manages multi-region load balancing
type LoadBalancer struct {
	router       *GeoRouter
	healthChecks map[string]HealthStatus
}

// HealthStatus represents the health of a region
type HealthStatus struct {
	RegionID  string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(router *GeoRouter) *LoadBalancer {
	return &LoadBalancer{
		router:       router,
		healthChecks: make(map[string]HealthStatus),
	}
}

// GetHealthyRegion returns a healthy region for routing
func (lb *LoadBalancer) GetHealthyRegion(tenantID uuid.UUID, location *Location) (Region, error) {
	region := lb.router.GetRegion(tenantID, location)

	// Check if region is healthy
	if status, exists := lb.healthChecks[region.ID]; exists && !status.Healthy {
		// Find backup region
		for _, r := range lb.router.regions {
			if r.ID != region.ID {
				if s, exists := lb.healthChecks[r.ID]; !exists || s.Healthy {
					return r, nil
				}
			}
		}
		return Region{}, fmt.Errorf("no healthy regions available")
	}

	return region, nil
}

// UpdateHealthCheck updates health status for a region
func (lb *LoadBalancer) UpdateHealthCheck(regionID string, healthy bool, latency time.Duration) {
	lb.healthChecks[regionID] = HealthStatus{
		RegionID:  regionID,
		Healthy:   healthy,
		Latency:   latency,
		CheckedAt: time.Now(),
	}
}

// RegionMiddleware handles region routing for HTTP requests
func RegionMiddleware(lb *LoadBalancer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract tenant ID from request
			tenantIDStr := r.Header.Get("X-Tenant-ID")
			if tenantIDStr == "" {
				tenantIDStr = "default"
			}

			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
			}

			// Get region
			region, err := lb.GetHealthyRegion(tenantID, nil)
			if err != nil {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			}

			// Add region info to context
			ctx := context.WithValue(r.Context(), "region", region)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
