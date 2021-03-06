package image // import "github.com/docker/docker/integration/image"

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/integration/internal/container"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/skip"
)

func TestRemoveImageOrphaning(t *testing.T) {
	skip.If(t, testEnv.DaemonInfo.OSType == "windows", "FIXME")
	defer setupTest(t)()
	ctx := context.Background()
	client := testEnv.APIClient()

	img := "test-container-orphaning"

	// Create a container from busybox, and commit a small change so we have a new image
	cID1 := container.Create(ctx, t, client, container.WithCmd(""))
	commitResp1, err := client.ContainerCommit(ctx, cID1, types.ContainerCommitOptions{
		Changes:   []string{`ENTRYPOINT ["true"]`},
		Reference: img,
	})
	assert.NilError(t, err)

	// verifies that reference now points to first image
	resp, _, err := client.ImageInspectWithRaw(ctx, img)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(resp.ID, commitResp1.ID))

	// Create a container from created image, and commit a small change with same reference name
	cID2 := container.Create(ctx, t, client, container.WithImage(img), container.WithCmd(""))
	commitResp2, err := client.ContainerCommit(ctx, cID2, types.ContainerCommitOptions{
		Changes:   []string{`LABEL Maintainer="Integration Tests"`},
		Reference: img,
	})
	assert.NilError(t, err)

	// verifies that reference now points to second image
	resp, _, err = client.ImageInspectWithRaw(ctx, img)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(resp.ID, commitResp2.ID))

	// try to remove the image, should not error out.
	_, err = client.ImageRemove(ctx, img, types.ImageRemoveOptions{})
	assert.NilError(t, err)

	// check if the first image is still there
	resp, _, err = client.ImageInspectWithRaw(ctx, commitResp1.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(resp.ID, commitResp1.ID))

	// check if the second image has been deleted
	_, _, err = client.ImageInspectWithRaw(ctx, commitResp2.ID)
	assert.Check(t, is.ErrorContains(err, "No such image:"))
}
