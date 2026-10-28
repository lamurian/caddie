package source

// limit keeps at most n jobs (all when n is 0), shared round-robin across
// feeds so that one long feed cannot crowd out the others (ADR 0018). A
// feed is one Greenhouse or Lever board, or one other source. The kept
// jobs stay in their original order.
func limit(jobs []Job, n int) []Job {
	if n <= 0 || len(jobs) <= n {
		return jobs
	}
	var order []string
	byFeed := map[string][]int{}
	for i, j := range jobs {
		k := feed(j)
		if _, ok := byFeed[k]; !ok {
			order = append(order, k)
		}
		byFeed[k] = append(byFeed[k], i)
	}
	keep := make([]bool, len(jobs))
	for kept, round := 0, 0; kept < n; round++ {
		for _, k := range order {
			if idx := byFeed[k]; round < len(idx) && kept < n {
				keep[idx[round]] = true
				kept++
			}
		}
	}
	out := make([]Job, 0, n)
	for i, j := range jobs {
		if keep[i] {
			out = append(out, j)
		}
	}
	return out
}

// feed names the list a job came from. Board jobs are told apart by
// company, which is the board's company name or, failing that, its slug.
func feed(j Job) string {
	if j.Source == Greenhouse || j.Source == Lever {
		return j.Source + "\x00" + j.Company
	}
	return j.Source
}
