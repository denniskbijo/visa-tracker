package eligibility

func formatPounds(p int64) string {
	if p == 0 {
		return "£0"
	}
	return "£" + formatCommas(p) + " /yr"
}

func formatPoundsPlain(p int64) string {
	return "£" + formatCommas(p)
}

func formatCommas(p int64) string {
	s := ""
	n := p
	for n > 0 {
		if s != "" {
			s = "," + s
		}
		chunk := n % 1000
		n /= 1000
		if n > 0 {
			if chunk < 10 {
				s = "00" + itoa(chunk) + s
			} else if chunk < 100 {
				s = "0" + itoa(chunk) + s
			} else {
				s = itoa(chunk) + s
			}
		} else {
			s = itoa(chunk) + s
		}
	}
	if s == "" {
		return "0"
	}
	return s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
