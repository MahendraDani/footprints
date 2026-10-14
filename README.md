## idea

> Explore personal browsing history through queries and interactive visualizations. The goal is to turn browser history from a collection of URLs into something we can explore, helping people understand their browsing patterns, revisit research trails, and discover connections between pages - chatgpt

bullshit, I know that having History database is power, but I don't know how to use it yet and in which ways? this doc describes things I've figured out so far and I'd like to use myself.

## Install

**macOS and Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/install.sh | sh
```

This puts the `footprints` binary in `~/.footprints/bin` and adds that folder to your `PATH`.
Open a new terminal (or run `source ~/.zshrc`) and check:

```bash
footprints --version
```

**Windows:** download `footprints_windows_amd64.zip` from the
[latest release](https://github.com/MahendraDani/footprints/releases/latest), unzip it, and put
`footprints.exe` somewhere on your `PATH`.

### Quick start

```bash
footprints init     # copy your Chrome history into ~/.footprints
footprints graph    # generate the visit graph and open it in your browser
```

### Update

footprints checks for new releases once a day and prints a notice when one is available. To upgrade:

```bash
footprints update
```

To turn the check off, set `FOOTPRINTS_NO_UPDATE_CHECK=1`.

### Install a specific version

```bash
curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/install.sh | sh
```

### Uninstall

```bash
rm ~/.footprints/bin/footprints
```

Then remove the `# footprints` PATH line from your shell config. Your data lives in `~/.footprints`, so delete that folder too if you want everything gone.

### Build from source

Requires Go 1.26+.

```bash
git clone https://github.com/MahendraDani/footprints.git
cd footprints
go build -o footprints .
```

## decisions

- **Interface:** CLI (def! I'm biased on using Go, but python's easier for me to write/read.)
- **Distribution:** A downloadable binary that users can run locally (i dunno if it's creepy but having a optional config to track analytics is good or bad? I feel like that cookie pop that annoys me on every sight. But without analytics how would I know how many people use my tool? at least knowing number of users is needed!).
- **Input data:** Chrome's History SQLite database. (yeah obviously! this was the starting point of this idea)
- **Visualizations:** Graphs are a core part of the product, not just supplementary reports. (bullshit! I just happened to observered the right thing, `visits` table self-references itself by storing `from_visit` and that gives me relations between diff page visit, then I prompted claude and it gave me the query and graph template. so yeah graphs are cool but only if it's a obsidian-level UX. that would be really crazy!)
- **Initial direction:** Explore browsing behavior through small experiments before committing to a fixed feature set. (why so serious?!)

## What we've proven so far (oh, really, huh?)

### Visit graph

**Problem:** Browser history is stored as records, making relationships between visits difficult to see.

**Approach:** Use the `visits` table's self-referencing relationship (`from_visit`) to reconstruct connections between visits, export the graph as JSON, and visualize it using HTML, CSS, and JavaScript.

**Result:** The prototype produces a visual representation of browsing activity that can be explored to understand navigation patterns.

**Status:** Working prototype.

## Potential features

These are hypotheses to investigate, not committed features.

| Feature | Question it could answer |
|---|---|
| Visit graph | How did I navigate between pages? |
| Recurring websites | Which domains do I repeatedly return to? |
| Research trails | Which pages did I visit while exploring a topic? |
| Search trails | What did I search for, and which pages did I visit afterward? |
| Browsing timelines | How does my browsing activity change over time? |
| Topic clusters | Which pages and topics appear related to each other? |
| Session exploration | What was I trying to accomplish during a browsing session? |


- research trail and search trail is same thing
- i think, as a user, just give me my graph, and I'll figure out the rest.
- browsing timeline, really? this is cool project, I'm not selling a SAAS built in 2018: here's is your dashboard with 1000 features. bullshit!
- topic cluster: bullshit! it's implicit in the graph, you open it and you'll already see everything
- session exploration: bullshit! bullshit! bullshit!

## Distribution and execution

**Proposed workflow:**

1. User downloads the binary.
2. User points it to their Chrome History database.
3. User selects or runs an analysis.
4. The CLI processes the data and produces a visualization.
5. The user explores the resulting graph or report.

**Open questions:**
- Should the CLI generate standalone HTML reports or run a local visualization server?
- Should users explicitly provide the database path, or should the tool discover it?
- Should visualizations be interactive by default?
- Which operating systems should the first release support?

nah, don't need to care much! I need to get the thing going ASAP. I'll iterate and build-up, but it needs to be public really really soon.

## Privacy and data handling

Browser history can contain sensitive personal information.

**Proposed default:** Process browsing data locally and avoid transmitting history to external services.

Questions to resolve:
- Is any telemetry necessary?
- What data, if any, should be collected to measure adoption?
- How should exported graphs and reports be protected?

ok, these are good questions, will follow-up!

## Next experiments

The next step is to test whether different representations reveal useful information, rather than immediately implementing every feature.

1. **Improve the visit graph:** Explore filtering, time ranges, and navigation paths.
2. **Test recurring-domain analysis:** Compare visit frequency with the number of distinct pages visited.
3. **Explore research trails:** Investigate whether visits, timestamps, referrers, and search terms can reveal coherent sequences of exploration.
4. **Evaluate the experience:** Determine which visualizations help answer questions that are difficult to answer using SQL alone.

For each experiment, document the question, SQL or data model, visualization, observations, and decision.

## Open product questions

- Who is the primary user: curious individuals, researchers, developers, or another audience?
- What is the central problem the product should solve?
- Which visualization is valuable enough to motivate someone to download and use the tool?
- What is the smallest useful first release?
- How will we measure whether users find the results useful?


NOOOOOOOOOO! i'm not thinking like a lousy PM at any cost! this is not a product! this is a cool fun project as I reverse engineer a software that everyone uses everyday. I'm not building a Saas in 2018!

