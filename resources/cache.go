package resources

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
)

type ResourceCache struct {
	path string
}

// Computes the cache key for the given Track instance.
func cacheKey(track *Track) string {
	hash := md5.New()
	io.WriteString(hash, track.Artist)

	if track.Album != "" {
		io.WriteString(hash, track.Album)
	} else {
		io.WriteString(hash, track.Title)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Returns the full path to the given resource. The path is computed and returned regardless of
// the existence of data at the location to where it points.
// This method returns an erro if it failed to make the directories leading to the path.
func (this *ResourceCache) resourcePath(track *Track, resourceType ResourceType) (string, error) {
	dir := fmt.Sprintf("%s/%s", this.path, resourceType)
	err := os.MkdirAll(dir, os.ModePerm|os.ModeDir)
	if err != nil {
		logger.Fatal("Failed to create directory %s: %s", dir, err)
		return "", err
	}
	return fmt.Sprintf("%s/%s", dir, cacheKey(track)), nil
}

func NewResourceCache() *ResourceCache {

	currentUser, err := user.Current()

	if err != nil {
		logger.Fatal("Failed to find the running user")
	}

	cachePath := currentUser.HomeDir + "/.cache/psmpc"

	logger.Info("Caching resources in %s", cachePath)
	return &ResourceCache{path: cachePath}
}

// Returns a local filesystem path to where the resource is cached
func (this *ResourceCache) Get(track *Track, resourceType ResourceType) (string, error) {
	return this.resourcePath(track, resourceType)
}

func (this *ResourceCache) Has(track *Track, resourceType ResourceType) bool {
	resourcePath, err := this.resourcePath(track, resourceType)
	if err != nil {
		logger.FatalError(err)
		return false
	}

	_, err = os.Stat(resourcePath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		logger.FatalError(err)
		return false
	}

	return true
}

func (this *ResourceCache) Set(track *Track, resourceType ResourceType, resourceStream io.ReadCloser) error {
	defer resourceStream.Close()
	resourcePath, err := this.resourcePath(track, resourceType)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resourceStream)
	if err != nil {
		logger.Warn("Failed to retrieve %s for %s: %s", resourceType, track, err)
		return err
	}

	return ioutil.WriteFile(resourcePath, data, os.ModePerm)
}

func (this *ResourceCache) Delete(track *Track, resourceType ResourceType) {
	// TODO
}
