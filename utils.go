package enforcer

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

func ruleExecEnv(ac Actor, act *Action) map[string]interface{} {
	d := map[string]interface{}{}
	if act != nil {
		d["event"] = mergeMap(act.Data, map[string]interface{}{"id": act.ID, "time": act.Time})
	}
	d["actor"] = mergeMap(ac.Attribs, map[string]interface{}{"id": ac.ID})
	return d
}

func mergeMap(m1, m2 map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}

func collectCampaignIDs(existing []Enrolment) []int {
	var res []int
	for _, enrolment := range existing {
		res = append(res, enrolment.CampaignID)
	}
	return res
}
