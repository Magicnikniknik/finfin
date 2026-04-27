package pricing

import (
	"sort"
)

func SelectBestMarginRule(rules []MarginRule, cmd CalculateQuoteCommand, side QuoteSide, volume string) (MarginRule, error) {
	eligible := make([]MarginRule, 0, len(rules))
	for _, r := range rules {
		if r.Side != side {
			continue
		}
		if !isVolumeBasisCompatible(cmd.InputMode, r.VolumeBasis) {
			continue
		}
		ok, err := isVolumeInRuleRange(volume, r)
		if err != nil {
			return MarginRule{}, err
		}
		if ok {
			eligible = append(eligible, r)
		}
	}
	if len(eligible) == 0 {
		return MarginRule{}, ErrNoMarginRuleFound
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		a, b := eligible[i], eligible[j]
		aOffice := a.OfficeID != nil
		bOffice := b.OfficeID != nil
		if aOffice != bOffice {
			return aOffice
		}
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		cmp, _ := compareDecimalStrings(a.MinVolume, b.MinVolume)
		if cmp != 0 {
			return cmp > 0
		}
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.After(b.CreatedAt)
		}
		return a.ID > b.ID
	})

	return eligible[0], nil
}

func isVolumeInRuleRange(volume string, rule MarginRule) (bool, error) {
	cmpMin, err := compareDecimalStrings(volume, rule.MinVolume)
	if err != nil {
		return false, err
	}
	if cmpMin < 0 {
		return false, nil
	}
	if rule.MaxVolume == nil {
		return true, nil
	}
	cmpMax, err := compareDecimalStrings(volume, *rule.MaxVolume)
	if err != nil {
		return false, err
	}
	return cmpMax <= 0, nil
}

func isVolumeBasisCompatible(inputMode QuoteInputMode, basis VolumeBasis) bool {
	switch basis {
	case "":
		// Backward compatibility for legacy rules without explicit volume_basis.
		return inputMode == InputModeGive
	case VolumeBasisGive:
		return inputMode == InputModeGive
	case VolumeBasisGet:
		return inputMode == InputModeGet
	case VolumeBasisQuoteNotional:
		return true
	default:
		return false
	}
}
