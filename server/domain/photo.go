package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Photo struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp    time.Time          `json:"timestamp" bson:"timestamp"`
	ImageType    string             `json:"image_type" bson:"image_type"`
	PresignedURL string             `json:"presigned_url" bson:",omitempty"`
	DeviceID     string             `json:"device_id" bson:"device_id"`
	Text         string             `json:"text" bson:"text"`

	// Medical Data Fields - flattened (each as separate column)
	UnitateMedicala        string    `json:"unitate_medicala" bson:"unitate_medicala"`
	AdresaUnitateMedicala  string    `json:"adresa_unitate_medicala" bson:"adresa_unitate_medicala"`
	TelefonUnitateMedicala string    `json:"telefon_unitate_medicala" bson:"telefon_unitate_medicala"`
	NumarFisa              string    `json:"numar_fisa" bson:"numar_fisa"`
	SocietateUnitate       string    `json:"societate_unitate" bson:"societate_unitate"`
	AdresaAngajator        string    `json:"adresa_angajator" bson:"adresa_angajator"`
	TelefonAngajator       string    `json:"telefon_angajator" bson:"telefon_angajator"`
	Nume                   string    `json:"nume" bson:"nume"`
	Prenume                string    `json:"prenume" bson:"prenume"`
	CNP                    string    `json:"cnp" bson:"cnp"`
	ProfesieFunctie        string    `json:"profesie_functie" bson:"profesie_functie"`
	LocDeMunca             string    `json:"loc_de_munca" bson:"loc_de_munca"`
	TipControl             string    `json:"tip_control" bson:"tip_control"`
	ControlAngajare        bool      `json:"control_angajare" bson:"control_angajare"`
	ControlPeriodic        bool      `json:"control_periodic" bson:"control_periodic"`
	ControlAdaptare        bool      `json:"control_adaptare" bson:"control_adaptare"`
	ControlReluare         bool      `json:"control_reluare" bson:"control_reluare"`
	ControlSupraveghere    bool      `json:"control_supraveghere" bson:"control_supraveghere"`
	ControlAlte            bool      `json:"control_alte" bson:"control_alte"`

	AvizMedical            string    `json:"aviz_medical" bson:"aviz_medical"`
	AvizApt                bool      `json:"aviz_apt" bson:"aviz_apt"`
	AvizAptConditionat     bool      `json:"aviz_apt_conditionat" bson:"aviz_apt_conditionat"`
	AvizInaptTemporar      bool      `json:"aviz_inapt_temporar" bson:"aviz_inapt_temporar"`
	AvizInapt              bool      `json:"aviz_inapt" bson:"aviz_inapt"`
	Recomandari            string    `json:"recomandari" bson:"recomandari"`
	Data                   time.Time `json:"data" bson:"data"`
	DataUrmExaminari       time.Time `json:"data_urm_examinari" bson:"data_urm_examinari"`
}

type PhotoRepository interface {
	GetPhotos(ctx context.Context, filters map[string]any) ([]*Photo, error)
	GetByID(ctx context.Context, id string) (*Photo, error)
	Save(ctx context.Context, photo *Photo) error
	UpdatePhotoTextAndMedicalData(ctx context.Context, id primitive.ObjectID, updates map[string]any) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context) (int64, error)
}

