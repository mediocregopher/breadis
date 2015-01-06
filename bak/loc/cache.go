package loc

import (
	"time"

	log "github.com/grooveshark/golib/gslog"

	"github.com/mediocregopher/breadis/config"
)

type minStack map[string]int

func (m minStack) add(k string, n, tokeep int) {
	if len(m) < tokeep {
		m[k] = n
		return
	}
	for mk, mn := range m {
		if n < mn {
			delete(m, mk)
			m[k] = n
			return
		}
	}
}

type cacheEntry struct {
	key, loc string
	hits     int
}

type cacheGet struct {
	key string
	ch  chan string
}

type cacheSet struct {
	key, loc string
}

var (
	getCh = make(chan *cacheGet)
	setCh = make(chan *cacheSet)
)

func init() {
	if config.CacheSize > 0 {
		go cacheSpin()
	}
}

func cacheSpin() {
	cache := map[string]*cacheEntry{}
	tick := time.Tick(5 * time.Minute)
	for {
		select {
		case g := <-getCh:
			if ent, ok := cache[g.key]; ok {
				ent.hits++
				g.ch <- ent.loc
			} else {
				g.ch <- ""
			}
		case s := <-setCh:
			cache[s.key] = &cacheEntry{s.key, s.loc, 0}
		case <-tick:
			log.Debug("Cleaning cache")
			if len(cache) <= config.CacheSize {
				break
			}
			ms := minStack{}
			for k, ent := range cache {
				ms.add(k, ent.hits, len(cache)-config.CacheSize)
			}
			for k := range ms {
				log.Debug("Deleting %s from cache", k)
				delete(cache, k)
			}
		}
	}
}

func getFromCache(key string) string {
	g := cacheGet{key, make(chan string)}
	// If there is any delay at all we bail, the cache could be garbage
	// collection which might take a non-trivial amount of time
	select {
	case getCh <- &g:
		return <-g.ch
	default:
		return ""
	}
}

func setInCache(key, loc string) {
	// If there is any delay at all we bail, the cache could be garbage
	// collection which might take a non-trivial amount of time
	select {
	case setCh <- &cacheSet{key, loc}:
	default:
	}
}
