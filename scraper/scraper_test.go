package scraper

import (
	"testing"
)

// Tests the synergy entry regex against a payload that mirrors a real OP.GG RSC chunk
// (escaped JSON inside a Next.js stream). Three entries: top-rank Garen, mid-rank Darius,
// and a champion with an apostrophe-stripped key (khazix) to confirm the [a-z0-9]+ class
// covers cleaned names.
func TestSynergyEntryPattern(t *testing.T) {
	payload := `something_unrelated":"foo",[{\"play\":13667,\"synergy_position\":\"TOP\",\"win_rate\":0.541158,\"pick_rate\":0.04666511878828438,\"synergy_champion_name\":\"Garen\",\"synergy_champion_image_url\":\"https://opgg-static.akamaized.net/meta/images/lol/16.8.1/champion/Garen.png\",\"synergy_champion_key\":\"garen\",\"tier_rank\":1},{\"play\":10963,\"synergy_position\":\"TOP\",\"win_rate\":0.515279,\"pick_rate\":0.0374,\"synergy_champion_name\":\"Darius\",\"synergy_champion_image_url\":\"https://opgg-static.akamaized.net/meta/images/lol/16.8.1/champion/Darius.png\",\"synergy_champion_key\":\"darius\",\"tier_rank\":2},{\"play\":8025,\"synergy_position\":\"JUNGLE\",\"win_rate\":0.5276,\"pick_rate\":0.027,\"synergy_champion_name\":\"Kha'Zix\",\"synergy_champion_image_url\":\"https://opgg-static.akamaized.net/meta/images/lol/16.8.1/champion/Khazix.png\",\"synergy_champion_key\":\"khazix\",\"tier_rank\":9}]`
	assertSynergyMatches(t, payload)

	plainPayload := `{"play":14431,"synergy_position":"SUPPORT","win_rate":0.565311,"pick_rate":0.26602392759046584,"synergy_champion_name":"Seraphine","synergy_champion_image_url":"https://opgg-static.akamaized.net/meta/images/lol/16.9.1/champion/Seraphine.png","synergy_champion_key":"seraphine","tier_rank":2}`
	plainMatches := synergyEntryPattern.FindAllStringSubmatch(plainPayload, -1)
	if len(plainMatches) != 1 {
		t.Fatalf("expected 1 plain-json entry, got %d", len(plainMatches))
	}
	if plainMatches[0][1] != "14431" || plainMatches[0][2] != "SUPPORT" || plainMatches[0][3] != "0.565311" || plainMatches[0][4] != "seraphine" || plainMatches[0][5] != "2" {
		t.Fatalf("unexpected plain-json match: %#v", plainMatches[0])
	}
}

func assertSynergyMatches(t *testing.T, payload string) {
	t.Helper()
	matches := synergyEntryPattern.FindAllStringSubmatch(payload, -1)
	if len(matches) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(matches))
	}

	type want struct {
		play, rank int
		role, key  string
		wr         string
	}
	wants := []want{
		{13667, 1, "TOP", "garen", "0.541158"},
		{10963, 2, "TOP", "darius", "0.515279"},
		{8025, 9, "JUNGLE", "khazix", "0.5276"},
	}

	for i, m := range matches {
		if m[1] != itoa(wants[i].play) {
			t.Errorf("entry %d: play=%s, want %d", i, m[1], wants[i].play)
		}
		if m[2] != wants[i].role {
			t.Errorf("entry %d: role=%s, want %s", i, m[2], wants[i].role)
		}
		if m[3] != wants[i].wr {
			t.Errorf("entry %d: wr=%s, want %s", i, m[3], wants[i].wr)
		}
		if m[4] != wants[i].key {
			t.Errorf("entry %d: key=%s, want %s", i, m[4], wants[i].key)
		}
		if m[5] != itoa(wants[i].rank) {
			t.Errorf("entry %d: rank=%s, want %d", i, m[5], wants[i].rank)
		}
	}
}

// itoa avoids strconv import in test.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
