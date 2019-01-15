package watcher

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/IBM/ubiquity/utils/logs"
)

/**
 * It is a simple watcher to watch a certain resource, if you want to watch more
 * that one resource or have complex process, use resource controller instead.
 */

// Watch watches a certain resource and call the handler when an event comes.
func Watch(watcher watch.Interface, handler cache.ResourceEventHandler, ctx context.Context, logger logs.Logger) error {
	var err error
	logger.Info("Start watching resource")

	objCache := make(map[string]interface{})

outerLoop:
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Added {
				name, err := cache.MetaNamespaceKeyFunc(event.Object)
				if err != nil {
					logger.Error(err.Error())
				} else {
					logger.Info("Resource created")
					objCache[name] = event.Object
					handler.OnAdd(event.Object)
				}

			} else if event.Type == watch.Modified {
				metadata, err := meta.Accessor(event.Object)
				if err != nil {
					logger.Error("Can not get resource metadata")
				} else {
					if metadata.GetDeletionTimestamp() == nil {
						logger.Info("Resource modified")
						name, _ := cache.MetaNamespaceKeyFunc(event.Object)
						handler.OnUpdate(objCache[name], event.Object)
						objCache[name] = event.Object
					}
				}
			} else if event.Type == watch.Deleted {
				logger.Info("Resource deleted")
				handler.OnDelete(event.Object)
				name, err := cache.MetaNamespaceKeyFunc(event.Object)
				if err == nil {
					delete(objCache, name)
				}
			}
		case <-ctx.Done():
			logger.Info("Shutting down the watcher")
			break outerLoop
		}
	}
	// stop watching this resource
	watcher.Stop()
	logger.Info("Stop watching resource")
	if err != nil {
		return err
	} else {
		return nil
	}
}
