package utils

func ConvertFloat64ToFloat32Embeddings(embeddings [][]float64) [][]float32 {
	float32Embeddings := make([][]float32, len(embeddings))
	for i, vec64 := range embeddings {
		vec32 := make([]float32, len(vec64))
		for j, v := range vec64 {
			vec32[j] = float32(v)
		}
		float32Embeddings[i] = vec32
	}
	return float32Embeddings
}

func ConvertFloat64ToFloat32Embedding(embedding []float64) []float32 {
	float32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		float32Embedding[i] = float32(v)
	}
	return float32Embedding
}
