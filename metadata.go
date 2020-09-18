package workfusion

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	URL      string `md:"url,required"` // The URL used to connect to the WorkFusion API
	Username string `md:"username"`     // The username used to connect to the WorkFusion API
	Password string `md:"password"`     // The password used to connect to the WorkFusion API
}

type Input struct {
	UUID string `md:"uuid"` // The UUID of the business process to copy and run
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"uuid": i.UUID,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {

	var err error
	i.UUID, err = coerce.ToParams(values["uuid"])
	if err != nil {
		return err
	}

	return nil
}

type Output struct {
	UUID string      `md:"uuid"` // The UUID of the new business process
	Data interface{} `md:"data"` // The final results data
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"uuid": o.UUID,
		"data": o.Data,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.UUID, err = coerce.ToInt(values["uuid"])
	if err != nil {
		return err
	}

	o.Data, _ = values["data"]

	return nil
}
