module tailscale.com

go 1.16

require (
	github.com/alexbrainman/sspi v0.0.0-20210105120005-909beea2cc74
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aws/aws-sdk-go v1.38.52
	github.com/coreos/go-iptables v0.6.0
	github.com/dapr/cli v0.9.0
	github.com/dapr/dapr v1.0.0-rc.1.0.20201202053523-152c218d89da
	github.com/dapr/dashboard v0.6.0
	github.com/frankban/quicktest v1.13.0
	github.com/gliderlabs/ssh v0.3.2
	github.com/go-multierror/multierror v1.0.2
	github.com/go-ole/go-ole v1.2.5
	github.com/godbus/dbus/v5 v5.0.4
	github.com/google/go-cmp v0.5.6
	github.com/google/goexpect v0.0.0-20210430020637-ab937bf7fd6f
	github.com/google/uuid v1.1.2
	github.com/goreleaser/nfpm v1.10.3
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/jsimonetti/rtnetlink v0.0.0-20210525051524-4cc836578190
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/klauspost/compress v1.12.2
	github.com/kr/pty v1.1.8
	github.com/mdlayher/netlink v1.4.1
	github.com/mdlayher/sdnotify v0.0.0-20210228150836-ea3ec207d697
	github.com/miekg/dns v1.1.42
	github.com/pborman/getopt v1.1.0
	github.com/peterbourgon/ff/v2 v2.0.0
	github.com/pkg/sftp v1.13.0
	github.com/spf13/viper v1.7.1
	github.com/tailscale/certstore v0.0.0-20210528134328-066c94b793d3
	github.com/tailscale/depaware v0.0.0-20201214215404-77d1e9757027
	github.com/tcnksm/go-httpstat v0.2.0
	github.com/toqueteos/webbrowser v1.2.0
	go4.org/mem v0.0.0-20201119185036-c04c5a6ff174
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6
	golang.org/x/tools v0.1.2
	golang.zx2c4.com/wireguard v0.0.0-20210525143454-64cb82f2b3f5
	golang.zx2c4.com/wireguard/windows v0.3.15-0.20210525143335-94c0476d63e3
	honnef.co/go/tools v0.1.4
	inet.af/netaddr v0.0.0-20210602152128-50f8686885e3
	inet.af/netstack v0.0.0-20210622165351-29b14ebc044e
	inet.af/peercred v0.0.0-20210318190834-4259e17bb763
	inet.af/wf v0.0.0-20210516214145-a5343001b756
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.2
	rsc.io/goversion v1.2.0
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client => github.com/kubernetes-client/go v0.0.0-20190928040339-c757968c4c36
