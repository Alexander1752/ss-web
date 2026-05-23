package utils

import (
	"strings"
	"testing"
)

func TestParseMedicalCertificate_ConditionalPaths(t *testing.T) {
	t.Run("returns nil for empty text", func(t *testing.T) {
		if got := ParseMedicalCertificate(""); got != nil {
			t.Fatalf("expected nil for empty text, got %+v", got)
		}
	})

	t.Run("returns nil for ocr failed", func(t *testing.T) {
		if got := ParseMedicalCertificate("OCR failed"); got != nil {
			t.Fatalf("expected nil for OCR failed, got %+v", got)
		}
	})

	t.Run("parses known fields and dates", func(t *testing.T) {
		input := "UNITATEA MEDICALA: Clinica Demo\nADRESA: Str. Test 1\nTEL: 021-000\n" +
			"FISA DE APTITUDINE NR. 77\n" +
			"Societate, unitate, etc. ACME Corp\nAdresa: Str. Ang 2\nTelefon: 031-111\n" +
			"NUME: Popescu\nPRENUME: Ion\nCNP: 1234567890123\n" +
			"Profesie / functie: Inginer\nLocul de munca: Lab\n" +
			"Data: 30/03/2026\nData urmatoarei examinari: 30/03/2027\n"

		data := ParseMedicalCertificate(input)
		if data == nil {
			t.Fatalf("expected parsed data, got nil")
		}
		if data.UnitateMedicala != "Clinica Demo" {
			t.Fatalf("expected medical unit Clinica Demo, got %q", data.UnitateMedicala)
		}
		if !strings.Contains(data.Nume, "Popescu") || !strings.Contains(data.Prenume, "Ion") {
			t.Fatalf("expected parsed person fields to include names, got nume=%q prenume=%q", data.Nume, data.Prenume)
		}
		if data.Data.IsZero() || data.DataUrmExaminari.IsZero() {
			t.Fatalf("expected parsed dates, got data=%v next=%v", data.Data, data.DataUrmExaminari)
		}
	})
}

func TestIsMedicalCertificate_KeywordThreshold(t *testing.T) {
	t.Run("false when fewer than two keywords", func(t *testing.T) {
		if IsMedicalCertificate("MEDICINA MUNCII only") {
			t.Fatalf("expected false for one keyword")
		}
	})

	t.Run("true when at least two keywords", func(t *testing.T) {
		if !IsMedicalCertificate("FISA DE APTITUDINE\nAVIZ MEDICAL") {
			t.Fatalf("expected true for two keywords")
		}
	})
}
