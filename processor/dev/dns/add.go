package dns

import (
	"fmt"
	"os"

	"github.com/nanobox-io/nanobox/models"
	"github.com/nanobox-io/nanobox/processor"
	"github.com/nanobox-io/nanobox/util"
	"github.com/nanobox-io/nanobox/util/config"
	"github.com/nanobox-io/nanobox/util/data"
	"github.com/nanobox-io/nanobox/util/dns"
)

// processDevDNSAdd ...
type processDevDNSAdd struct {
	control processor.ProcessControl
}

//
func init() {
	processor.Register("dev_dns_add", devDNSAddFunc)
}

//
func devDNSAddFunc(control processor.ProcessControl) (processor.Processor, error) {
	return processDevDNSAdd{control: control}, nil
}

//
func (devDNSAdd processDevDNSAdd) Results() processor.ProcessControl {
	return devDNSAdd.control
}

//
func (devDNSAdd processDevDNSAdd) Process() error {

	//
	name := devDNSAdd.control.Meta["name"]

	//
	app := models.App{}
	data.Get("apps", config.AppName(), &app)

	//
	preview := dns.Entry(app.DevIP, name, "preview")
	dev := dns.Entry(app.DevIP, name, "dev")

	// if the entry doesnt exist just return
	if !dns.Exists(preview) && !dns.Exists(dev) {
		return nil
	}

	// This process requires root, check to see if we're the root user. If not, we
	// need to run a hidden command as sudo that will just call this function again.
	// Thus, the subprocess will be running as root
	if os.Geteuid() != 0 {

		// get the original nanobox executable
		nanobox := os.Args[0]

		// call 'dev dns add' with the original path (ultimately leads right back here)
		cmd := fmt.Sprintf("%s dev dns add %s", nanobox, name)

		// if the sudo'ed subprocess fails, we need to return error to stop the process
		fmt.Println("Admin privileges are required to add DNS entries to your hosts file, your password may be requested...")
		if err := util.PrivilegeExec(cmd); err != nil {
			return err
		}

		// the subprocess exited successfully, so we can short-circuit here
		return nil
	}

	// add the 'preview' DNS entry into the /etc/hosts file
	if err := dns.Add(preview); err != nil {
		return err
	}

	// add the 'dev' DNS entry into the /etc/hosts file
	if err := dns.Add(dev); err != nil {
		return err
	}

	return nil
}