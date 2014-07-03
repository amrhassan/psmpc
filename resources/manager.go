package resources

import (
	"container/list"
	"errors"
	"io"
	"log"
)

type ResourceManager struct {
	providers map[ResourceType]*list.List
	cache     *ResourceCache
}

var (
	ResourceNotFound            = errors.New("The requested resource could not be found")
	ResourceCouldNotBeRetrieved = errors.New("The requested resource could not be retrieved")
)

func NewResourceManager() *ResourceManager {
	manager := &ResourceManager{
		providers: make(map[ResourceType]*list.List),
		cache:     NewResourceCache(),
	}

	// Register your providers here
	manager.registerProvider(newLastFMAlbumArtProvider())

	return manager
}

func (this *ResourceManager) registerProvider(resourceProvider ResourceProvider) {
	_, exists := this.providers[resourceProvider.Type()]
	if !exists {
		this.providers[resourceProvider.Type()] = list.New()
	}

	this.providers[resourceProvider.Type()].PushBack(resourceProvider)
}

func (this *ResourceManager) getResourceFromProvider(track *Track, resourceType ResourceType) (io.ReadCloser, error) {
	list := this.providers[resourceType]
	for provider := list.Front(); provider != nil; provider = provider.Next() {
		resource, err := provider.Value.(ResourceProvider).GetResource(track)

		if err != nil {
			log.Printf("Failed to get %s from %s for %s", resourceType, provider.Value, track)
			return nil, err
		}

		return resource, nil
	}

	return nil, ResourceNotFound
}

// Returns a local filesystem path to a file where the resource is fully available, or returns one
// of ResourceNotFound or ResourceCouldNotBeRetrieved for errors.
func (this *ResourceManager) GetResourceAsFilePath(track *Track, resourceType ResourceType) (string, error) {

	if this.cache.Has(track, resourceType) {
		resourcePath, err := this.cache.Get(track, resourceType)

		if err != nil {
			log.Printf("Failed to retrieve %s for %s from cache", resourceType, track)
			this.cache.Delete(track, resourceType)
		} else {
			return resourcePath, nil
		}
	}

	resourceStream, err := this.getResourceFromProvider(track, resourceType)

	if err != nil {
		log.Printf("Failed to get resource %s for %s: %s", resourceType, track, err)
		return "", ResourceNotFound
	}

	err = this.cache.Set(track, resourceType, resourceStream)

	if err != nil {
		log.Println("Failed to retrieve the resource %s for %s: %s", resourceType, track, err)
		return "", ResourceCouldNotBeRetrieved
	}

	return this.cache.Get(track, resourceType)
}
