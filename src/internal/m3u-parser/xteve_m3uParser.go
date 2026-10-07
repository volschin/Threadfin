package m3u

// Keep the historical entry point on the same parser as large playlists.
func makeInterfaceFromM3UOriginal(byteStream []byte) ([]interface{}, error) {
	return MakeInterfaceFromM3UOptimized(byteStream)
}
