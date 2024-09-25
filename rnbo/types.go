package rnbo

type Set struct {
	Id                  int
	Name                string
	Filename            string
	Runner_rnbo_version string
	Created_at          string
	Meta                string
}

type SetConnection struct {
	Id                    int
	Set_Id                int
	Source_name           string
	Source_instance_index int
	Source_port_name      string
	Sink_name             string
	Sink_instance_index   int
	Sink_port_name        string
}

type SetPatcherInstance struct {
	Id                 int
	Patcher_id         int
	Set_id             int
	Set_instance_index int
	Config             string
}

type SetPreset struct {
	Id                 int
	Patcher_id         int
	Set_id             int
	Set_instance_index int
	Name               string
	Content            string
	Initial            int
	Created_at         string
	Updated_at         string
}
