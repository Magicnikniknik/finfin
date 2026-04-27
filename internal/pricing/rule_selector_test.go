package pricing

import (
	"testing"
	"time"
)

func TestSelectBestMarginRule_PriorityAndOffice(t *testing.T) {
	office := "o1"
	rules := []MarginRule{
		{ID: "g1", OfficeID: nil, Side: SideSell, MinVolume: "0", MaxVolume: nil, Priority: 200, CreatedAt: time.Unix(1, 0)},
		{ID: "o1", OfficeID: &office, Side: SideSell, MinVolume: "0", MaxVolume: nil, Priority: 100, CreatedAt: time.Unix(2, 0)},
	}
	got, err := SelectBestMarginRule(rules, CalculateQuoteCommand{InputMode: InputModeGive}, SideSell, "50")
	if err != nil {
		t.Fatalf("SelectBestMarginRule err = %v", err)
	}
	if got.ID != "o1" {
		t.Fatalf("winner = %s, want office rule", got.ID)
	}
}

func TestSelectBestMarginRule_NoRule(t *testing.T) {
	_, err := SelectBestMarginRule(nil, CalculateQuoteCommand{}, SideSell, "1")
	if err != ErrNoMarginRuleFound {
		t.Fatalf("err = %v, want %v", err, ErrNoMarginRuleFound)
	}
}
