package entity

type Categoria string

const (
	Politica    Categoria = "politica"
	Economia    Categoria = "economia"
	Inseguridad Categoria = "inseguridad"
	Sensible    Categoria = "sensible"
	Otros       Categoria = "otros"
)

type Clasificacion struct {
	Category         Categoria `bson:"category" json:"category"`
	SensitivityLevel int64     `bson:"sensitivityLevel" json:"sensitivityLevel"`
}
