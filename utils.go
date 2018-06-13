package dlrouter

func GetReversedBytes(str []byte) []byte {
	bytes := []byte(str)
	for st, end := 0, len(bytes)-1; st < end; st, end = st+1, end-1 {
		bytes[st], bytes[end] = bytes[end], bytes[st]
	}
	return bytes
}

func RemoveDuplicates(slice []*Target) []*Target {
	for i := 0; i < len(slice); i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i].Value == slice[j].Value {
				slice = append(slice[:j], slice[j+1:]...)
				j--
			}
		}
	}
	return slice
}
