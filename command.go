package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	packagecloud "github.com/atotto/packagecloud/api/v1"
	"github.com/google/subcommands"
)

type pushCommand struct {
}

func (*pushCommand) Name() string     { return "push" }
func (*pushCommand) Synopsis() string { return "pushing a package" }
func (*pushCommand) Usage() string {
	return `push name/repo/distro/version filepath

example:
    packagecloud push example-user/example-repository/ubuntu/xenial /tmp/example.deb
`
}
func (p *pushCommand) SetFlags(f *flag.FlagSet) {}
func (p *pushCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	repos, distro, version, ok := splitPackageTarget(f.Arg(0))
	if !ok {
		return subcommands.ExitUsageError
	}
	fpath := f.Arg(1)
	if err := packagecloud.PushPackage(ctx, repos, distro, version, fpath); err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

type deleteCommand struct {
}

func (*deleteCommand) Name() string     { return "yank" }
func (*deleteCommand) Synopsis() string { return "deleting a package" }
func (*deleteCommand) Usage() string {
	return `yank name/repo/distro/version filepath

example:
    packagecloud yank example-user/example-repository/ubuntu/xenial example_1.0.1-1_amd64.deb
`
}
func (p *deleteCommand) SetFlags(f *flag.FlagSet) {}
func (p *deleteCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	repos, distro, version, ok := splitPackageTarget(f.Arg(0))
	if !ok {
		return subcommands.ExitUsageError
	}
	fpath := f.Arg(1)
	if err := packagecloud.DeletePackage(ctx, repos, distro, version, fpath); err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

type promoteCommand struct {
}

func (*promoteCommand) Name() string     { return "promote" }
func (*promoteCommand) Synopsis() string { return "promote package" }
func (*promoteCommand) Usage() string {
	return `promote name/src_repo/distro/version filepath name/dst_repo

example:
    packagecloud promote example-user/repo1/ubuntu/xenial example_1.0-1_amd64.deb example-user/repo2
`
}
func (p *promoteCommand) SetFlags(f *flag.FlagSet) {}
func (p *promoteCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	srcRepos, distro, version, ok := splitPackageTarget(f.Arg(0))
	if !ok {
		return subcommands.ExitUsageError
	}
	fpath := f.Arg(1)
	dstRepos := f.Arg(2)
	if err := packagecloud.PromotePackage(ctx, dstRepos, srcRepos, distro, version, fpath); err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func splitPackageTarget(target string) (repos, distro, version string, ok bool) {
	ss := strings.SplitN(target, "/", 4)
	if len(ss) != 4 {
		ok = false
		return
	}
	repos = fmt.Sprintf("%s/%s", ss[0], ss[1])
	distro = ss[2]
	version = ss[3]
	ok = true
	return
}
