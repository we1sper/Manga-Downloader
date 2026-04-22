# Requirements for the ManHuaGui WebUI

## Overview

Design and develop a web UI that supports manga search, download task submission, and task status inquiry.

## Techniques

Utilize vanilla JavaScript and CSS exclusively, with no external frameworks or libraries.

## Style

The web UI should be concise. You should refer to GitHub pages designed in a minimalist black, grey and white colour scheme.

You are encouraged to incorporate appropriate animations to enrich the page and enhance user engagement.

For instance, a background particle system can be implemented. Particles are assigned initial velocities upon page loading. They interact with one another: whenever three particles come close together, they form triangular connections, which dissolve once the particles drift apart.

## Detail

You need to meet the following requirements:

### 1. A search input field with a button is located at the top of the page

- This input only accepts positive integers, which correspond to manga ID
- Pressing the Enter key triggers the search directly, no button click required
- The search request API is defined as `Query Manga` in `apis.md`

### 2. Search result area

- Displays search results accurately with all necessary information fully presented
- Manga cover is displayed as an image
- Manga chapters are presented in table format

### 3. Download task submission

- A checkbox is added to each row of the result table; checking a checkbox selects the corresponding chapter for download submission
- A select-all checkbox button is provided
- The download request API is defined as `Download Chapters` in `apis.md`
- Search results only include chapter URLs formatted as `https://www.manhuagui.com/comic/<mid>/<cid>.html`. Thus, `mid` and `cid` must be extracted before submitting download tasks

### 4. Download task status

- The task status query API is defined as `Query Download Records` in `apis.md`
- Scheduled jobs will periodically call this API to fetch task status
- This module is divided into two tabs: ongoing tasks and historical tasks, both presented in table format
- A progress bar is displayed for active tasks (tasks that are neither completed nor failed)
- Failed tasks feature a retry button in their corresponding table row, which invokes the `Download Chapters` API
- A bulk retry button is also provided for batch task retry
- Toast notifications will appear at the bottom-right corner once tasks complete successfully or fail

### 5. Additional

- Table height should be restricted, and overflow content should support scroll viewing
- Add a spinner animation to indicate loading status during search and submit tasks processes
- Add a clear icon to the search input field, which appears when there is text in the input. Clicking this icon clears the input field and any displayed search results
- Support sorting chapters by title or page count in ascending or descending order by clicking the corresponding table header
