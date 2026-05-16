# 🕷️ SiteMapr

> **A bold, fast, and highly concurrent sitemap crawler & SEO parser built in Go.**

SiteMapr rips through massive XML sitemaps, unrolls standard pages, and concurrently fetches essential SEO markers (`Title`, `H1`, `MetaDescription`, `StatusCode`) without breaking a sweat. It natively bypasses basic bot protections through user-agent rotation and exports clean, structured datasets in your requested format.

---

## 🚀 Features

- **Concurrent Execution Engine**: Operates with custom-defined worker limits utilizing Go routines and semaphore channels. No race conditions, strict concurrency limits.
- **Sitemap Orhchestration**: Differentiates between XML index nodes and regular page endpoints to dynamically build worklists.
- **Smart Delays & Bypasses**: Randomly rotates common cross-platform User-Agents to mimic human navigation while strictly enforcing HTTP execution timeouts.
- **Intercli & Output Flexibility**: Ask-on-run terminal prompts. Request JSON or CSV output arrays to instantly visualize bulk SEO tags.

## 🛠️ Get Started

Clone the beast and navigate to your directory:

```bash
git clone https://github.com/RudraPratapDev/SiteMapr.git
cd SiteMapr
```

Resolve any dependencies:
```bash
make tidy
```

Unleash the crawler:
```bash
make run
```
You will be prompted to supply a starting URL, maximum concurrency limit, and output destination (CSV/JSON). The crawler handles the rest.

## 📂 Architecture

SiteMapr relies on a modular separation of concerns.
- `/scraper`: Handles network logic, XML node extraction, and concurrency mechanisms.
- `/exporter`: Consumes structured datasets and serializes them to localized file caches. 

## ⚖️ License
MIT. Go ahead—break it, bend it, mod it. 
