package workfusion

import (
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

const (
	BaseUrl  = "https://wf-tibpoc-984-wfaw-10049-workfusion-lb1.workfusion.com/workfusion/api"
	Username = ""
	Password = ""
	BPId     = "4eb1e6e5-f722-45e9-af49-e5380cf14003"
)

func TestRegister(t *testing.T) {

	ref := activity.GetRef(&Activity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func TestSettings(t *testing.T) {

	// valid settings
	settings := &Settings{
		URL:      BaseUrl,
		Username: Username,
		Password: Password,
	}

	iCtx := test.NewActivityInitContext(settings, nil)
	_, err := New(iCtx)
	assert.Nil(t, err)

	// No URL
	settings = &Settings{
		URL:      "",
		Username: Username,
		Password: Password,
	}

	iCtx = test.NewActivityInitContext(settings, nil)
	_, err = New(iCtx)
	assert.NotNil(t, err)

	// Bad URL
	settings = &Settings{
		URL:      "https://tibco.com",
		Username: Username,
		Password: Password,
	}

	iCtx = test.NewActivityInitContext(settings, nil)
	_, err = New(iCtx)
	assert.NotNil(t, err)

	// No user
	settings = &Settings{
		URL:      BaseUrl,
		Username: "",
		Password: "BadPassword",
	}

	iCtx = test.NewActivityInitContext(settings, nil)
	_, err = New(iCtx)
	assert.NotNil(t, err)

	// No password
	settings = &Settings{
		URL:      BaseUrl,
		Username: Username,
		Password: "",
	}

	iCtx = test.NewActivityInitContext(settings, nil)
	_, err = New(iCtx)
	assert.NotNil(t, err)

	// Bad creds
	settings = &Settings{
		URL:      BaseUrl,
		Username: "BadUsername",
		Password: "BadPassword",
	}

	iCtx = test.NewActivityInitContext(settings, nil)
	_, err = New(iCtx)
	assert.NotNil(t, err)
}

func TestEvalSuccess(t *testing.T) {

	settings := &Settings{
		URL:      BaseUrl,
		Username: Username,
		Password: Password,
	}

	mf := mapper.NewFactory(resolve.GetBasicResolver())
	iCtx := test.NewActivityInitContext(settings, mf)
	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())

	//setup attrs
	tc.SetInput("uuid", BPId)

	//eval
	act.Eval(tc)
	assert.NotNil(t, tc.GetOutput("uuid"))
	assert.NotNil(t, tc.GetOutput("data"))
}

func TestEvalBadBPId(t *testing.T) {

	settings := &Settings{
		URL:      BaseUrl,
		Username: Username,
		Password: Password,
	}

	mf := mapper.NewFactory(resolve.GetBasicResolver())
	iCtx := test.NewActivityInitContext(settings, mf)
	act, err := New(iCtx)
	assert.Nil(t, err)

	tc := test.NewActivityContext(act.Metadata())

	//setup attrs
	tc.SetInput("uuid", "bad1e6e5-f722-45e9-af49-e5380cf14003")

	//eval
	act.Eval(tc)
	assert.NotNil(t, tc.GetOutput("uuid"))
	assert.NotNil(t, tc.GetOutput("data"))
}
