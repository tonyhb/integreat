package registry

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/integreat/modules"
	itypes "github.com/docker/integreat/types"
	"github.com/docker/integreat/util"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/distribution/xfer"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/types"
	"github.com/docker/libtrust"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func init() {
	modules.Register("registry", itypes.ModuleCreator(NewSuite))
}

func NewSuite(opts itypes.ModuleOpts) (itypes.Module, error) {
	cfg, ok := opts.Config["registry"]
	if !ok {
		return nil, fmt.Errorf("registry config not found")
	}
	url, _ := url.Parse(cfg["host"].(string))
	key, _ := libtrust.GenerateECP256PrivateKey()
	return &Registry{
		url:    url,
		rand:   opts.Rand,
		logger: opts.Logger,
		key:    key,
	}, nil
}

type Registry struct {
	rand   *rand.Rand
	logger *logrus.Logger
	url    *url.URL

	key libtrust.PrivateKey
}

func (r *Registry) GetCommand(cmd string) (itypes.TestCommand, error) {
	return modules.GetCommand(r, cmd)
}

func (r *Registry) PushRandomImage(a itypes.TestArgs) (itypes.TestResult, error) {
	if users, ok := a["createUsers"]; ok {
		for _, user := range users.([]itypes.TestResult) {
			if err := r.pushRandomImage(user["name"].(string), "test"); err != nil {
				return nil, err
			}
		}
	}
	return itypes.TestResult{}, nil
}

func (r *Registry) pushRandomImage(namespace, name string) error {
	ctx := context.Background()
	tag := util.RandomString(r.rand, 10)

	repo, err := r.getRepo(ctx, namespace, name, "password")
	if err != nil {
		return err
	}

	// Create each random layer
	layers := []xfer.UploadDescriptor{
		&v2LayerPush{
			log:         r.logger,
			rand:        r.rand,
			layerNumber: 0,
			size:        67108864,
			repo:        repo,
		},
	}
	lum := xfer.NewLayerUploadManager(1)
	if err = lum.Upload(ctx, layers, new(BlankProgress)); err != nil {
		return err
	}

	// Attempt V2 manifest first
	// Create the manifest for this image
	builder := schema2.NewManifestBuilder(repo.Blobs(ctx), []byte("{}"))
	for _, i := range layers {
		if err := builder.AppendReference(i.(*v2LayerPush)); err != nil {
			return err
		}
	}
	manifest, err := builder.Build(ctx)
	if err != nil {
		return err
	}
	manSvc, _ := repo.Manifests(ctx)
	putOptions := []distribution.ManifestServiceOption{distribution.WithTag(tag)}
	if _, err = manSvc.Put(ctx, manifest, putOptions...); err != nil {

		diffids := []string{}
		for _, i := range layers {
			diffids = append(diffids, i.(*v2LayerPush).Descriptor().Digest.String())
		}

		config := map[string]interface{}{
			"history": []map[string]interface{}{
				{
					"author": "integreat",
				},
			},
			"rootfs": map[string]interface{}{
				"diff_ids": diffids,
			},
		}

		configByt, _ := json.Marshal(config)

		// Fall back to V1 manifest (DTR 2.0)
		manifestRef, err := reference.WithTag(repo.Named(), tag)
		if err != nil {
			return err
		}
		builder = schema1.NewConfigManifestBuilder(repo.Blobs(ctx), r.key, manifestRef, configByt)
		for _, i := range layers {
			builder.AppendReference(i.(*v2LayerPush))
		}
		manifest, err := builder.Build(ctx)
		if err != nil {
			return fmt.Errorf("error building manifest: %s", err)
		}
		fmt.Println(manifest)
		if _, err = manSvc.Put(ctx, manifest, putOptions...); err != nil {
			return fmt.Errorf("error saving manifest: %s", err)
		}
	}

	return nil
}

func (r *Registry) getRepo(ctx context.Context, user, repoName, pass string) (distribution.Repository, error) {
	direct := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	base := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                direct.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:   true,
	}

	modifiers := registry.DockerHeaders("integreat", http.Header{})
	authTransport := transport.NewTransport(base, modifiers...)
	challengeManager, _, err := registry.PingV2Registry(r.url, authTransport)

	// Set up auth config
	authCfg := &types.AuthConfig{
		Username:      user,
		Password:      pass,
		ServerAddress: r.url.Host,
	}
	creds := registry.NewStaticCredentialStore(authCfg)
	tokenHandlerOptions := auth.TokenHandlerOptions{
		Transport:   authTransport,
		Credentials: creds,
		Scopes: []auth.Scope{
			auth.RepositoryScope{
				Repository: user + "/" + repoName,
				Actions:    []string{"push", "pull"},
			},
		},
		ClientID: registry.AuthClientID,
	}
	tokenHandler := auth.NewTokenHandlerWithOptions(tokenHandlerOptions)
	basicHandler := auth.NewBasicHandler(creds)
	modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))

	repoNameRef, err := reference.ParseNamed(user + "/" + repoName)
	tr := transport.NewTransport(base, modifiers...)
	repo, err := client.NewRepository(ctx, repoNameRef, r.url.String(), tr)
	return repo, err
}

type BlankProgress struct{}

func (b BlankProgress) WriteProgress(p progress.Progress) error {
	return nil
}
