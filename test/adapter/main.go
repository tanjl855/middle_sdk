package adapter

type Cm interface {
	getLength(float64) float64
}

type M interface {
	getLength(float64) float64
}

func NewM() M {
	return &getLengthM{}
}

type getLengthM struct{}

func (*getLengthM) getLength(cm float64) float64 {
	return cm / 10
}

func NewCm() Cm {
	return &getLengthCm{}
}

type getLengthCm struct{}

func (a *getLengthCm) getLength(m float64) float64 {
	return m * 10
}

// 适配器
type LengthAdapter interface {
	getLength(string, float64) float64
}

var _ LengthAdapter = &getLengthAdapter{}

func NewLengthAdapter() LengthAdapter {
	return &getLengthAdapter{}
}

type getLengthAdapter struct{}

func (*getLengthAdapter) getLength(isType string, into float64) float64 {
	if isType == "m" {
		return NewM().getLength(into)
	}

	return NewCm().getLength(into)
}
