package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	packagecloud "github.com/atotto/packagecloud/api/v1"
	"github.com/google/subcommands"
)

type commandBase struct {
	name         string
	synopsis     string
	usage        string
	examples     []string
	setFlagsFunc func(f *flag.FlagSet)
	executeFunc  func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus
}

func (c *commandBase) Name() string { return c.name }
func (c *commandBase) Synopsis() string {
	return fmt.Sprintf("packagecloud %s", c.synopsis)
}

func (c *commandBase) Usage() string {
	w := bytes.NewBufferString(c.usage)
	fmt.Fprintln(w)
	if len(c.examples) > 0 {
		fmt.Fprintln(w, "\nexample:")
		for _, ex := range c.examples {
			fmt.Fprintf(w, "    %s\n", ex)
		}
	}
	return w.String()
}

func (c *commandBase) SetFlags(f *flag.FlagSet) {
	if c.setFlagsFunc == nil {
		return
	}
	c.setFlagsFunc(f)
}

func (c *commandBase) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	return c.executeFunc(ctx, f, args)
}

var pushPackageCommand = &commandBase{
	"push",
	"pushing a package",
	"push name/repo/distro/version filepath",
	[]string{"packagecloud push example-user/example-repository/ubuntu/xenial /tmp/example.deb"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		repos, distro, version, n := splitPackageTarget(f.Arg(0))
		if n != 4 {
			return subcommands.ExitUsageError
		}
		fpath := f.Arg(1)
		if err := packagecloud.PushPackage(ctx, repos, distro, version, fpath); err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}

		return subcommands.ExitSuccess
	},
}

var searchPackageCommand = &commandBase{
	"list",
	"list package",
	"list name/repo query [version]",
	[]string{"packagecloud list example-user/example-repository example 1.0.0"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		repos, distro, _, n := splitPackageTarget(f.Arg(0))
		if n < 2 {
			return subcommands.ExitUsageError
		}
		query := f.Arg(1)
		details, err := packagecloud.SearchPackage(ctx, repos, distro, query, "")
		if err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}
		version := f.Arg(2)
		for _, detail := range details {
			if version == "" || detail.Version == version {
				fmt.Printf("%s %s %s %s\n", detail.DistroVersion, detail.Filename, detail.Name, detail.Version)
			}
		}

		return subcommands.ExitSuccess
	},
}

var pullPackageCommand = &commandBase{
	"pull",
	"pull package",
	"pull name/repo[/distro/version] filename",
	[]string{"packagecloud pull example-user/example-repository example_1.0.0_all.deb"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		repos, distro, version, n := splitPackageTarget(f.Arg(0))
		if n < 2 {
			return subcommands.ExitUsageError
		}
		query := f.Arg(1)
		details, err := packagecloud.SearchPackage(ctx, repos, distro, query, "")
		if err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}
		for _, detail := range details {
			if version == "" || detail.Version == version {
				fmt.Printf("%s %s %s %s\n", detail.Filename, detail.Name, detail.Version, detail.DistroVersion)
				f, err := os.OpenFile(detail.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					log.Println(err)
					return subcommands.ExitFailure
				}
				defer f.Close()
				resp, err := http.Get(detail.DownloadURL)
				if err != nil {
					log.Println(err)
					return subcommands.ExitFailure
				}
				defer resp.Body.Close()
				io.Copy(f, resp.Body)
			}
		}

		return subcommands.ExitSuccess
	},
}

var deletePackageCommand = &commandBase{
	"rm",
	"deleting a package",
	"rm name/repo/distro/version filepath",
	[]string{"packagecloud rm example-user/example-repository/ubuntu/xenial example_1.0.1-1_amd64.deb"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		repos, distro, version, n := splitPackageTarget(f.Arg(0))
		if n != 4 {
			return subcommands.ExitUsageError
		}
		fpath := f.Arg(1)
		if err := packagecloud.DeletePackage(ctx, repos, distro, version, fpath); err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}

		return subcommands.ExitSuccess
	},
}

var promotePackageCommand = &commandBase{
	"promote",
	"promote package",
	"promote name/src_repo/distro/version filepath name/dst_repo",
	[]string{"packagecloud promote example-user/repo1/ubuntu/xenial example_1.0-1_amd64.deb example-user/repo2"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		srcRepos, distro, version, n := splitPackageTarget(f.Arg(0))
		if n != 4 {
			return subcommands.ExitUsageError
		}
		fpath := f.Arg(1)
		dstRepos := f.Arg(2)
		if err := packagecloud.PromotePackage(ctx, dstRepos, srcRepos, distro, version, fpath); err != nil {
			log.Println(err)
			return subcommands.ExitFailure
		}

		return subcommands.ExitSuccess
	},
}

func splitPackageTarget(target string) (repos, distro, version string, n int) {
	ss := strings.SplitN(target, "/", 4)
	n = len(ss)
	if n >= 2 {
		repos = fmt.Sprintf("%s/%s", ss[0], ss[1])
	}
	if n >= 3 {
		distro = ss[2]
	}
	if n >= 4 {
		version = ss[3]
	}
	return
}

var helpDistroCommand = &commandBase{
	"distro",
	"list supported distributions",
	"distro [deb/py]",
	[]string{"packagecloud distro", "packagecloud distro deb", "packagecloud distro deb ubuntu", "packagecloud distro | jq .deb"},
	nil,
	func(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
		var v interface{}
		switch typ := f.Arg(0); typ {
		case "deb", "debian":
			if name := f.Arg(1); name != "" {
				for _, distros := range packagecloud.GetDistributions().Deb {
					if distros.IndexName == name {
						v = distros.Versions
						break
					}
				}
			} else {
				v = packagecloud.GetDistributions().Deb
			}

		case "py", "python":
			v = packagecloud.GetDistributions().Py
		case "":
			v = packagecloud.GetDistributions()
		default:
			log.Printf("not supported type:%s", typ)
			return subcommands.ExitUsageError
		}
		if err := json.NewEncoder(os.Stdout).Encode(v); err != nil {
			return subcommands.ExitFailure
		}
		return subcommands.ExitSuccess
	},
}
