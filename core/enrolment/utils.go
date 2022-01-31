package enrolment

func contains(arr []string, item string) bool {
	for _, s := range arr {
		if s == item {
			return true
		}
	}
	return false
}

func filterByStatus(arr []Enrolment, status []string) []Enrolment {
	if len(status) == 0 {
		return arr
	}

	var res []Enrolment
	for _, enr := range arr {
		enr.setStatus()
		if contains(status, enr.Status) {
			res = append(res, enr)
		}
	}
	return res
}
