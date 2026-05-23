package utils

import (
	"regexp"
	"strings"
	"time"
)

// MedicalData reprezinta datele structurate extrase din fisa de aptitudine medicala
type MedicalData struct {
	// Header - Unitatea Medicala
	UnitateMedicala        string `json:"unitate_medicala" bson:"unitate_medicala"`                 // UNITATEA MEDICALA
	AdresaUnitateMedicala  string `json:"adresa_unitate_medicala" bson:"adresa_unitate_medicala"`   // ADRESA (sus)
	TelefonUnitateMedicala string `json:"telefon_unitate_medicala" bson:"telefon_unitate_medicala"` // TEL / FAX (sus)

	// Header - Tip Fisa
	NumarFisa string `json:"numar_fisa" bson:"numar_fisa"` // FISA DE APTITUDINE NR.

	// Sectiune Angajator / Institutie
	SocietateUnitate string `json:"societate_unitate" bson:"societate_unitate"` // Societate, unitate, etc.
	AdresaAngajator  string `json:"adresa_angajator" bson:"adresa_angajator"`   // Adresa (jos)
	TelefonAngajator string `json:"telefon_angajator" bson:"telefon_angajator"` // Telefon / Fax (jos)

	// Date Personale Angajat
	Nume    string `json:"nume" bson:"nume"`       // NUME
	Prenume string `json:"prenume" bson:"prenume"` // PRENUME
	CNP     string `json:"cnp" bson:"cnp"`         // CNP

	// Date Profesionale
	ProfesieFunctie string `json:"profesie_functie" bson:"profesie_functie"` // Profesie / functie
	LocDeMunca      string `json:"loc_de_munca" bson:"loc_de_munca"`         // Locul de munca

	// Date Medicale
	TipControl       string    `json:"tip_control" bson:"tip_control"`               // Angajare, Control medical periodic, etc.
	ControlAngajare  bool      `json:"control_angajare" bson:"control_angajare"`     // Angajare (checkbox)
	ControlPeriodic  bool      `json:"control_periodic" bson:"control_periodic"`     // Control medical periodic (checkbox)
	ControlAdaptare  bool      `json:"control_adaptare" bson:"control_adaptare"`     // Adaptare (checkbox)
	ControlReluare   bool      `json:"control_reluare" bson:"control_reluare"`       // Reluarea muncii (checkbox)
	ControlSupraveghere bool   `json:"control_supraveghere" bson:"control_supraveghere"` // Supraveghere speciala (checkbox)
	ControlAlte      bool      `json:"control_alte" bson:"control_alte"`             // Alte (checkbox)

	AvizMedical      string    `json:"aviz_medical" bson:"aviz_medical"`             // APT, APT CONDITIONAT, etc.
	AvizApt          bool      `json:"aviz_apt" bson:"aviz_apt"`                     // APT (checkbox)
	AvizAptConditionat bool    `json:"aviz_apt_conditionat" bson:"aviz_apt_conditionat"` // APT CONDITIONAT (checkbox)
	AvizInaptTemporar bool     `json:"aviz_inapt_temporar" bson:"aviz_inapt_temporar"`   // INAPT TEMPORAR (checkbox)
	AvizInapt        bool      `json:"aviz_inapt" bson:"aviz_inapt"`                 // INAPT (checkbox)

	Recomandari      string    `json:"recomandari" bson:"recomandari"`               // RECOMANDARI field
	Data             time.Time `json:"data" bson:"data"`                             // Data
	DataUrmExaminari time.Time `json:"data_urm_examinari" bson:"data_urm_examinari"` // Data urmatoarei examinari
}

// ParseMedicalCertificate extrage datele structurate din textul OCR al unei fise de aptitudine medicala romanesti
func ParseMedicalCertificate(ocrText string) *MedicalData {
	if ocrText == "" || ocrText == "OCR failed" {
		return nil
	}

	data := &MedicalData{}

	// --- Header Section (Unitatea Medicala) ---
	// UNITATEA MEDICALA
	data.UnitateMedicala = extractField(ocrText, `UNITATEA\s+MEDICALA:\s*([^\n]+)`)
	if data.UnitateMedicala == "" {
		// Fallback for messy OCR
		data.UnitateMedicala = extractField(ocrText, `MEDICALA:\s*([^\n]+)`)
	}

	// Because "Adresa" appears twice, we try to capture the first occurrence for the Medical Unit
	// Typically at the top of the document.
	// We'll use a split strategy: try to split the document roughly in half or by "Societate" keyword.
	parts := strings.Split(ocrText, "Societate")
	topPart := parts[0]
	bottomPart := ocrText
	if len(parts) > 1 {
		bottomPart = "Societate" + parts[1] // Include "Societate" back for regex matching
	}

	// Adresa Unitate Medicala (from top part)
	data.AdresaUnitateMedicala = extractField(topPart, `ADRESA:\s*([^\n]+)`)

	// Telefon Unitate Medicala (from top part)
	data.TelefonUnitateMedicala = extractField(topPart, `TEL:\s*([^\n]+)`)

	// Numar Fisa
	data.NumarFisa = extractField(ocrText, `FISA\s+DE\s+APTITUDINE\s+NR\.?\s*(\d+)`)

	// --- Employer Section (Societate, unitate) ---
	data.NumarFisa = extractField(ocrText, `(?i)FISA\s+DE\s+APTITUDINE\s+NR[\.:]?\s*(\d+)`)

	// --- Employer Section (from bottom part) ---
	// Societate, unitate, etc.
	data.SocietateUnitate = extractMultilineField(bottomPart, `(?i)Soci[ec]tate,?\s*unitate,?\s*(?:etc[\.:]?)?\s*([^\n]+(?:\n[^\n]+)?)`)
	// Try explicit university match if above failed or just as fallback
	if data.SocietateUnitate == "" {
		data.SocietateUnitate = extractField(bottomPart, `(?i)(UNIVERSITATEA\s+(?:NATIONALA\s+DE\s+STIINTA\s+SI\s+TEHNOLOGIE\s+)?POLITEHNICA\s+(?:DIN\s+)?[A-Z]+)`)
	}

	// Adresa Angajator
	// Note: In bottom part, ADRESA appears again.
	data.AdresaAngajator = extractField(bottomPart, `(?i)Adresa[:;]?\s*([^\n]+)`)

	// Telefon/Fax Angajator
	data.TelefonAngajator = extractField(bottomPart, `(?i)(?:Telefon|Fax)[:;]?\s*([^\n]+)`)

	// --- Personal Data ---
	data.Nume = extractField(ocrText, `(?i)NUME[:;]?\s*([A-Za-z\s]+)`)
	data.Prenume = extractField(ocrText, `(?i)PRENUME[:;]?\s*([A-Za-z\s]+)`)
	data.CNP = extractField(ocrText, `(?i)CNP[:;]?\s*(\d+)`)

	// --- Professional Data ---
	data.ProfesieFunctie = extractField(ocrText, `(?i)Profesie\s*[\/\|]\s*functie[:;]?\s*([^\n]+)`)
	
	data.LocDeMunca = extractField(ocrText, `(?i)Locul?\s+de\s+munca[:;]?\s*([^\n]+)`)

	// --- Medical Data ---
	// --- Medical Data ---
	// Helper function for gap analysis
	isBoxChecked := func(text string, currentLabel, nextLabel string) bool {
		// Find current label index
		curIdx := strings.Index(text, currentLabel)
		if curIdx == -1 {
			// Try fuzzy match if exact fail? For now, assume labels are found or logic fails gracefully
			// If label not found, return false
			return false
		}
		
		// Find next label index, searching AFTER current label
		// If nextLabel is empty, search until end of line/string segment
		searchStart := curIdx + len(currentLabel)
		var gapText string
		if nextLabel != "" {
			nextIdx := strings.Index(text[searchStart:], nextLabel)
			if nextIdx == -1 {
				// Next label not found, look at a reasonable window (e.g., 20 chars)
				end := searchStart + 20
				if end > len(text) {
					end = len(text)
				}
				gapText = text[searchStart:end]
			} else {
				gapText = text[searchStart : searchStart+nextIdx]
			}
		} else {
			// No next label, look at rest of line or small window
			end := searchStart + 20
			if end > len(text) {
				end = len(text)
			}
			gapText = text[searchStart:end]
		}

		// Check for "Empty Box" patterns in the gap
		// Patterns: [], [ ], [-], [[]
		emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
		if emptyBoxRegex.MatchString(gapText) {
			return false // Found an empty box
		}
		
		// If no empty box is found, AND the neighbors are close enough (gap not huge), assume checked
		// But verify we don't just have whitespace/newlines
		// Actually, if OCR eats the [X], we might just have spaces.
		return true
	}

	// Normalize text for easier matching (handle duplication if present, just take first occurrence)
	// The OCR might output: "Angajare ... Ane [] Angajare ... Ane []"
	// We'll work on the first occurrence of the START of the sequence.
	
	// Tip Control Row
	// Labels: Angajare, Control, Adaptare, Reluarea/Roluarca, Supraveghere, Alte/Ane
	// We need to handle fuzzy labels.
	// We'll replace fuzzy labels with standard ones in a temp string for easier indexing.
	
	// Work on topPart as it contains the control checkboxes
	tempTop := topPart
	tempTop = strings.ReplaceAll(tempTop, "Roluarca", "Reluarea")
	tempTop = strings.ReplaceAll(tempTop, "Ane", "Alte")
	
	// Handle "Control" carefully because "Control medical" might be split
	// Just searching for "Control" is fine as long as it's the specific header one.
	// "Angajare" should precede it.
	
	// Locate the start of the row to avoid false matches elsewhere
	rowStart := strings.Index(tempTop, "Angajare")
	if rowStart != -1 {
		rowText := tempTop[rowStart:] // Work from Angajare onwards
		
		data.ControlAngajare = isBoxChecked(rowText, "Angajare", "Control")
		data.ControlPeriodic = isBoxChecked(rowText, "Control", "Adaptare")
		data.ControlAdaptare = isBoxChecked(rowText, "Adaptare", "Reluarea")
		data.ControlReluare = isBoxChecked(rowText, "Reluarea", "Supraveghere")
		data.ControlSupraveghere = isBoxChecked(rowText, "Supraveghere", "Alte")
		data.ControlAlte = isBoxChecked(rowText, "Alte", "")
	}

	if data.ControlAngajare {
		data.TipControl = "Angajare"
	} else if data.ControlPeriodic || data.ControlAdaptare {
		if data.ControlAdaptare {
			data.TipControl = "Adaptare"
		} else {
			data.TipControl = "Control medical periodic"
		}
	} else if data.ControlReluare {
		data.TipControl = "Reluarea muncii"
	} else if data.ControlSupraveghere {
		data.TipControl = "Supraveghere speciala"
	}

	// Aviz Medical (Bottom checkboxes)
	// Labels: APT, APT CONDITIONAT, INAPT TEMPORAR, INAPT
	// These are in ocrText (likely bottom part).
	
	// Heuristic for APT: The OCR sometimes messes up "APT" vs "APT CONDITIONAT".
	// "APT" is a prefix of "APT CONDITIONAT".
	// GAP Analysis:
	// APT -> APT CONDITIONAT
	// APT CONDITIONAT -> INAPT TEMPORAR
	// INAPT TEMPORAR -> INAPT
	// INAPT -> End
	
	// Normalize some potential OCR errors for labels
	tempBottom := bottomPart
	tempBottom = strings.ReplaceAll(tempBottom, "ApT", "APT")
	tempBottom = strings.ReplaceAll(tempBottom, "aerconpmonat", "APT CONDITIONAT") // Extreme fuzzy fix based on logs? Maybe risky.
	// Looking at log: "aerconpmonat [ ApTCONDITIONAT []"
	// It seems "ApTCONDITIONAT" is found.
	
	// Locate start of Aviz section
	avizStart := strings.Index(tempBottom, "AVIZ MEDICAL")
	if avizStart != -1 {
		avizText := tempBottom[avizStart:]
		
		// Check APT. Gap matches between "Top APT" and "APT CONDITIONAT"
		// Wait, OCR: "APT: :" ... "ApTCONDITIONAT"
		// Regex index search might be safer than exact string replace for weird garbage.
		// "APT" followed by "APT CONDITIONAT" ?
		
		// Let's rely on finding "APT" (with colon usually) and "APT CONDITIONAT".
		// We use a custom parser for this row since labels overlap.
		
		// Find "APT" that is NOT "APT CONDITIONAT"
		// Using regex to find indices is better.
		
		aptIdx := -1
		aptCondIdx := -1
		aptCondEndIdx := -1
		inaptTempIdx := -1
		inaptIdx := -1
		
		// Find APT: (Look for APT followed by non-alpha)
		aptRegex := regexp.MustCompile(`APT[:;\s]+`)
		loc := aptRegex.FindStringIndex(avizText)
		if loc != nil {
			aptIdx = loc[0]
			// Use end of match to ensure we don't include the colon in the search if it was part of the match
			// But wait, if regex includes the colon, we want loc[1]
			aptIdx = loc[1] // Re-assign to end of match for gap start
		} else {
			// Try finding just APT
			aptRegex = regexp.MustCompile(`APT`)
			loc = aptRegex.FindStringIndex(avizText)
			if loc != nil {
				aptIdx = loc[1]
			}
		}

		// Find APT CONDITIONAT (or just CONDITIONAT to be safe)
		// Use Conditionat as the anchor
		aptCondRegex := regexp.MustCompile(`(?i)CONDITIONAT`) // Simplified from APT CONDITIONAT
		loc = aptCondRegex.FindStringIndex(avizText)
		if loc != nil {
			// We want the start of the word CONDITIONAT, but check if APT precedes it
			// Actually, just finding CONDITIONAT is enough for the End boundary of the gap
			aptCondIdx = loc[0]
			aptCondEndIdx = loc[1]
		}
		
		if aptIdx != -1 && aptCondIdx != -1 && aptIdx < aptCondIdx {
			// Analysis for APT
			// Gap between end of APT label and start of CONDITIONAT label
			gap := avizText[aptIdx : aptCondIdx]
			emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
			data.AvizApt = !emptyBoxRegex.MatchString(gap)
		}
		
		// Find INAPT TEMPORAR
		inaptTempRegex := regexp.MustCompile(`(?i)INAPT\s*TEMPORAR`)
		loc = inaptTempRegex.FindStringIndex(avizText)
		if loc != nil {
			inaptTempIdx = loc[0]
		}
		
		if aptCondEndIdx != -1 {
			end := inaptTempIdx
			if end == -1 { end = len(avizText) } // Fallback
			if end > aptCondEndIdx {
				gap := avizText[aptCondEndIdx : end]
				emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
				data.AvizAptConditionat = !emptyBoxRegex.MatchString(gap)
			}
		}

		// Find INAPT (last one)
		// Careful not to match INAPT inside INAPT TEMPORAR
		// Search AFTER INAPT TEMPORAR if found
		startSearch := 0
		if inaptTempIdx != -1 {
			startSearch = inaptTempIdx + len("INAPT TEMPORAR")
			
			// Check INAPT TEMPORAR status here
			// Gap is between INAPT TEMPORAR and INAPT
			// Find INAPT
			inaptRegex := regexp.MustCompile(`(?i)INAPT[:\s]+`)
			loc = inaptRegex.FindStringIndex(avizText[startSearch:])
			if loc != nil {
				inaptIdx = startSearch + loc[0]
				
				// Status for Inapt Temporar
				gap := avizText[startSearch : inaptIdx]
				emptyBoxRegex := regexp.MustCompile(`\[\s*[\[\]\-\|]?\s*\]`)
				data.AvizInaptTemporar = !emptyBoxRegex.MatchString(gap)
				
				// Status for Inapt
				// Look ahead a bit
				end := inaptIdx + 15
				if end > len(avizText) { end = len(avizText) }
				gapLast := avizText[inaptIdx+len("INAPT") : end]
				data.AvizInapt = !emptyBoxRegex.MatchString(gapLast)
				
			} else {
				// INAPT not found? weird.
				// Maybe use previous logic for fallback
			}
		}
	}

	if data.AvizApt {
		data.AvizMedical = "APT"
	} else if data.AvizAptConditionat {
		data.AvizMedical = "APT CONDITIONAT"
	} else if data.AvizInaptTemporar {
		data.AvizMedical = "INAPT TEMPORAR"
	} else if data.AvizInapt {
		data.AvizMedical = "INAPT"
	}

	// Dates
	// Data: ... (usually left bottom)
	// Relaxed regex for date (allows space around / or -)
	dateStr := extractField(bottomPart, `(?i)Data[:;]?\s*(\d{2}[\.\/\-]\d{2}[\.\/\-]\d{4})`)
	if dateStr != "" {
		// Replace . or - with / for parsing
		normalizedDate := strings.ReplaceAll(dateStr, ".", "/")
		normalizedDate = strings.ReplaceAll(normalizedDate, "-", "/")
		if t, err := time.Parse("02/01/2006", normalizedDate); err == nil {
			data.Data = t
		}
	}

	// Data urmatoarei examinari
	nextDateStr := extractField(bottomPart, `(?i)Data\s+urmatoarei\s+examinari[:;]?\s*(\d{2}[\.\/\-]\d{2}[\.\/\-]\d{4})`)
	if nextDateStr != "" {
		normalizedDate := strings.ReplaceAll(nextDateStr, ".", "/")
		normalizedDate = strings.ReplaceAll(normalizedDate, "-", "/")
		if t, err := time.Parse("02/01/2006", normalizedDate); err == nil {
			data.DataUrmExaminari = t
		}
	}
	
	return data
}

// extractField extrage un camp folosind un pattern regex
func extractField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractMultilineField extrage campuri care pot fi pe mai multe linii
func extractMultilineField(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		// Curata rezultatul
		result := strings.TrimSpace(matches[1])
		// Inlocuieste spatii multiple cu un singur spatiu
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		return result
	}
	return ""
}

// containsChecked verifica daca un camp are X (este bifat)
func containsChecked(text, fieldName string) bool {
	// Cauta pattern-uri ca "Angajare []", "Angajare [X]", "Angajare X"
	// The OCR might read checked boxes as [X], X, [x], x, or even just a weird symbol inside.
	// We'll look for the field name followed by a box-like structure containing X.
	
	// Complex regex to catch:
	// FieldName ... [ X ]
	// FieldName ... X
	// [ X ] FieldName
	
	// Simplest robust check: FieldName followed closely by X or [X]
	// \s* means optional whitespace
	// (?: ... ) is non-capturing group
	// \[? ... \]? matches optional brackets
	// [Xx] matches X or x
	
	// Case 1: Label then Box/Mark (e.g., "APT: [X]")
	patternRight := fieldName + `[:\s\._-]*\[?\s*[Xx]\s*\]?`
	if regexp.MustCompile(patternRight).MatchString(text) {
		return true
	}
	
	// Case 2: Box/Mark then Label (e.g., "[X] Adaptare")
	patternLeft := `\[?\s*[Xx]\s*\]?[:\s\._-]*` + fieldName
	if regexp.MustCompile(patternLeft).MatchString(text) {
		return true
	}

	return false
}

// IsMedicalCertificate verifica daca textul OCR este dintr-o fisa de aptitudine medicala
func IsMedicalCertificate(ocrText string) bool {
	keywords := []string{
		"MEDICINA MUNCII",
		"FISA DE APTITUDINE",
		"AVIZ MEDICAL",
		"APT",
	}
	
	matchCount := 0
	for _, keyword := range keywords {
		if strings.Contains(ocrText, keyword) {
			matchCount++
		}
	}
	
	// Daca gaseste cel putin 2 cuvinte cheie, este fisa medicala
	return matchCount >= 2
}
