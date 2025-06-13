# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based web scraper that continuously collects League of Legends champion matchup statistics from OP.GG and stores them in MongoDB. The scraper runs as a daemon, executing daily to fetch win rates and games played data for all champions across different positions and competitive tiers.

## Common Commands

### Running the scraper
```bash
go run main.go
```

### Building the project
```bash
go build -o opggscraper
```

### Managing dependencies
```bash
go mod download  # Download dependencies
go mod tidy      # Clean up dependencies
```

### MongoDB operations
```bash
# Connect to MongoDB (assuming local instance)
mongo mongodb://localhost:27017

# View scraped data
use live-lol-esports-stats
db.ranked_stats.find()
```

### Testing (no tests currently exist)
```bash
go test ./...    # Run all tests when they exist
```

## Code Architecture

The codebase follows a modular structure with database integration:

- **main.go**: Entry point that orchestrates the continuous scraping process
  - Implements 24-hour ticker for daily execution
  - Handles graceful shutdown with context cancellation (SIGINT/SIGTERM)
  - Manages MongoDB connection lifecycle
  - Performs validation before saving data
  - Scrapes multiple tiers: emerald_plus, diamond_plus, master_plus, grandmaster, challenger

- **db/mongo.go**: MongoDB database operations
  - Manages connection with timeout and ping verification
  - `SaveChampionStats()`: Upserts champion data based on champion/patch/tier
  - Uses database: "live-lol-esports-stats", collection: "ranked_stats"

- **model/models.go**: Data structures
  - `RankedChampionStats`: Main structure for MongoDB storage with champion data, patch, tier, and timestamp
  - `Matchup`: Stores win rate and games played statistics
  - `Position`: Enum for champion positions (Top, Jungle, Mid, Adc, Support)

- **scraper/scraper.go**: Web scraping logic using Colly framework
  - `GetChampionNames()`: Fetches all champion names from OP.GG
  - `GetChampionMatchups()`: Scrapes matchup data for all positions
  - `GetChampionMatchupsByPosition()`: Scrapes matchup data for specific position

- **utils/patch.go**: Patch version management
  - `GetLatestPatchVersion()`: Fetches latest patch from Riot's API
  - `FormatPatchVersion()`: Formats patch for database (15.07 → 15.7)
  - `FormatPatchVersionForOpGG()`: Formats patch for OP.GG URLs (15.7 → 15.07)

- **utils/utils.go**: General utilities
  - `CleanChampionName()`: Normalizes champion names for URL usage
  - `SaveJSON()`: Legacy JSON file saving (still present but not used)

## Key Implementation Details

1. **Continuous Operation**: Runs indefinitely with daily scraping cycles
2. **Dynamic Patch Detection**: Automatically fetches latest patch from Riot API
3. **Rate Limiting**: 
   - 15 minutes between tiers (currently set to 1 minute for testing)
   - 2 seconds between champions
4. **Validation**: Validates scraped data before saving (checks Ezreal vs Jinx ADC matchup)
5. **Context-aware Cancellation**: Properly handles shutdown at any point in the scraping process
6. **MongoDB Upsert**: Updates existing records or inserts new ones based on champion/patch/tier
7. **URL Construction**: `https://www.op.gg/champions/{championName}/counters/{position}?region=global&tier={tier}&patch={patch}`

## Recent Improvements

1. **Memory Efficiency**: Champions are now saved individually as they're scraped (no more in-memory accumulation)
2. **Immediate Results**: Data is available in MongoDB as soon as each champion is scraped
3. **Flexible Validation**: Uses `ValidateChampionData()` to check for any valid winrate (%) instead of specific matchup
4. **Fixed Memory Leak**: Each scraping operation now uses a fresh Colly collector to prevent callback accumulation

## Development Notes

- No existing tests - consider adding tests for scraper logic, database operations, and validation
- No linting configuration - consider adding `golangci-lint` for code quality
- The `.gitignore` excludes `*.json` files (legacy from previous version)
- MongoDB connection string is hardcoded in main.go:20
- All dependencies are managed through Go modules (go.mod)