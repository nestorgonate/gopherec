package entity

type Categoria string

const (
	Politica    Categoria = "Politica"
	Economica   Categoria = "Economica"
	Inseguridad Categoria = "Inseguridad"
	Sensible    Categoria = "Sensible"
	Otros       Categoria = "Otros"
)

type Clasificacion struct {
	Category         Categoria `bson:"category" json:"category"`
	SensitivityLevel int       `bson:"sensitivityLevel" json:"sensitivityLevel"`
}
