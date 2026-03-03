package domain

// NearbyWorker represents a worker found in a spatial proximity query.
// Used to build the weight matrix for Hungarian matching.
type NearbyWorker struct {
	WorkerID       string   `json:"worker_id"`
	Name           string   `json:"name"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	DistanceMeters float64  `json:"distance_meters"`
	H3Index        string   `json:"h3_index,omitempty"`
	RatingAvg      float64  `json:"rating_avg"`
	Skills         []string `json:"skills"`
	Tier           string   `json:"tier"`
}

// NearbyWorkersResponse wraps the nearby worker query result.
type NearbyWorkersResponse struct {
	Workers      []NearbyWorker `json:"workers"`
	QueryLat     float64        `json:"query_lat"`
	QueryLon     float64        `json:"query_lon"`
	RadiusMeters float64        `json:"radius_meters"`
	Count        int            `json:"count"`
}
