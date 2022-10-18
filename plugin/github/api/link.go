package api

type Link struct {
	Url string
	Rel string
}

type Links []Link

func (ls Links) FindByRel(rel string) (*Link, bool) {
	for _, link := range ls {
		if link.Rel == rel {
			return &link, true
		}
	}

	return nil, false
}

func ParseLinkHeader(s string) (links Links, err error) {
	sc := NewScanner(s)

	for !sc.AtEnd() {
		var captured []string

		captured, err = sc.Parse(
			ExpectRune{'<'},
			CaptureUpToRune{'>'},
			ExpectRune{';'},
			ConsumeWhitespace{},
			ExpectString{"rel=\""},
			CaptureUpToRune{'"'},
			ConsumeWhitespace{},
		)

		if err != nil {
			return links, err
		}

		links = append(links, Link{
			captured[0],
			captured[1],
		})

		if !sc.AtEnd() {
			_, err = sc.Parse(ExpectRune{','}, ConsumeWhitespace{})
			if err != nil {
				return
			}
		}
	}

	return links, nil
}
