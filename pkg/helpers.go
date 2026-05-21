package pkg

import "strconv"

func ParseInt(v string) int {
	i, _ := strconv.Atoi(v)
	return i
}

func GetOr(row []string, idx int) string {
	if len(row) > idx {
		return row[idx]
	}
	return ""
}

