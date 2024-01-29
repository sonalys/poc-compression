package segments

// func Test_GetCompressionGain(t *testing.T) {
// 	t.Run("masked segment", func(t *testing.T) {
// 		seg := NewMaskedSegment([]byte{255, 254, 244}, 1)
// 		originalSize := 3
// 		compressedSize := len(seg.Encode())
// 		require.Equal(t, compressedSize, seg.getCompressedSize())
// 		gains := seg.GetCompressionGains()
// 		require.Equal(t, originalSize-compressedSize, gains)
// 	})

// 	t.Run("repeat segment", func(t *testing.T) {
// 		seg := NewRepeatSegment(50, 1, 0, 100)
// 		originalSize := 100
// 		compressedSize := len(seg.Encode())
// 		require.Equal(t, compressedSize, seg.getCompressedSize())
// 		gains := seg.GetCompressionGains()
// 		require.Equal(t, originalSize-compressedSize, gains)
// 	})

// 	t.Run("group segment", func(t *testing.T) {
// 		seg := NewGroupSegment([]byte{1, 2}, 0, 100)
// 		originalSize := 4
// 		compressedSize := len(seg.Encode())
// 		require.Equal(t, compressedSize, seg.getCompressedSize())
// 		gains := seg.GetCompressionGains()
// 		require.Equal(t, originalSize-compressedSize, gains)
// 	})

// 	t.Run("group segment", func(t *testing.T) {
// 		seg := NewGroupSegment([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0, 100, 200)
// 		originalSize := 30
// 		compressedSize := len(seg.Encode())
// 		require.Equal(t, compressedSize, seg.getCompressedSize())
// 		gains := seg.GetCompressionGains()
// 		require.Equal(t, originalSize-compressedSize, gains)
// 	})
// }
