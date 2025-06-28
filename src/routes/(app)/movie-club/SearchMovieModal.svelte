<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { onMount } from "svelte";
	import type { Content } from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import axios from "axios";

	const dispatch = createEventDispatcher<{
		movieSelected: { content: Content; reason: string };
		close: void;
	}>();

	let searchQuery = "";
	let searchResults: Content[] = [];
	let selectedMovie: Content | null = null;
	let reason = "";
	let searching = false;
	let searchTimeout: NodeJS.Timeout;

	onMount(() => {
		// Focus search input
		const searchInput = document.getElementById("movie-search-input");
		if (searchInput) {
			searchInput.focus();
		}
	});

	async function searchMovies(query: string) {
		if (!query.trim()) {
			searchResults = [];
			return;
		}

		searching = true;
		try {
			const response = await axios.get(`/content/search/movie`, {
				params: {
					query: query
				}
			});
			searchResults = response.data.results || [];
		} catch (err) {
			console.error("Search failed:", err);
			searchResults = [];
		} finally {
			searching = false;
		}
	}

	function handleSearchInput() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			searchMovies(searchQuery);
		}, 300);
	}

	function selectMovie(movie: Content) {
		selectedMovie = movie;
		reason = "";
	}

	function submitNomination() {
		if (!selectedMovie) return;

		dispatch("movieSelected", {
			content: selectedMovie,
			reason: reason.trim()
		});
	}

	function goBack() {
		selectedMovie = null;
		reason = "";
	}
</script>

<div class="search-modal">
	<div class="modal-header">
		<h3>
			{#if selectedMovie}
				Nominate Movie
			{:else}
				Search for a Movie
			{/if}
		</h3>
		<button class="close-btn" on:click={() => dispatch("close")}>
			<Icon icon="close" />
		</button>
	</div>

	{#if !selectedMovie}
		<!-- Search Phase -->
		<div class="search-section">
			<div class="search-input-container">
				<Icon icon="search" />
				<input
					id="movie-search-input"
					type="text"
					placeholder="Search for movies..."
					bind:value={searchQuery}
					on:input={handleSearchInput}
				/>
			</div>

			{#if searching}
				<div class="searching">
					<Spinner />
					<p>Searching...</p>
				</div>
			{:else if searchResults.length > 0}
				<div class="search-results">
					{#each searchResults as movie}
						<div class="movie-result" on:click={() => selectMovie(movie)}>
							<div class="poster-small">
								<Poster content={movie} showRating={false} />
							</div>
							<div class="movie-info">
								<h4>{movie.title}</h4>
								<p class="release-year">
									{movie.release_date ? new Date(movie.release_date).getFullYear() : "Unknown"}
								</p>
								{#if movie.overview}
									<p class="overview">{movie.overview.substring(0, 150)}...</p>
								{/if}
							</div>
							<Icon icon="chevron" />
						</div>
					{/each}
				</div>
			{:else if searchQuery.trim()}
				<div class="no-results">
					<Icon icon="search" />
					<p>No movies found for "{searchQuery}"</p>
				</div>
			{:else}
				<div class="search-prompt">
					<Icon icon="film" />
					<p>Start typing to search for movies</p>
				</div>
			{/if}
		</div>
	{:else}
		<!-- Nomination Phase -->
		<div class="nomination-section">
			<button class="back-btn" on:click={goBack}>
				<Icon icon="arrow" />
				Back to Search
			</button>

			<div class="selected-movie">
				<div class="poster-large">
					<Poster content={selectedMovie} showRating={false} />
				</div>
				<div class="movie-details">
					<h4>{selectedMovie.title}</h4>
					<p class="release-year">
						{selectedMovie.release_date ? new Date(selectedMovie.release_date).getFullYear() : "Unknown"}
					</p>
					{#if selectedMovie.overview}
						<p class="overview">{selectedMovie.overview}</p>
					{/if}
				</div>
			</div>

			<div class="reason-section">
				<label for="reason">Why are you nominating this movie? (Optional)</label>
				<textarea
					id="reason"
					placeholder="Share why you think this would be a great choice..."
					bind:value={reason}
					maxlength="500"
				></textarea>
				<small>{reason.length}/500 characters</small>
			</div>

			<div class="action-buttons">
				<button class="cancel-btn" on:click={() => dispatch("close")}>
					Cancel
				</button>
				<button class="submit-btn" on:click={submitNomination}>
					<Icon icon="add" />
					Nominate This Movie
				</button>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.search-modal {
		width: 100%;
		max-width: 600px;
		max-height: 80vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background: var(--background);
		border-radius: 12px;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid var(--border);

		h3 {
			margin: 0;
			color: var(--text);
		}

		.close-btn {
			background: none;
			border: none;
			padding: 0.5rem;
			cursor: pointer;
			color: var(--text-muted);
			border-radius: 4px;
			transition: all 0.2s ease;

			&:hover {
				color: var(--text);
				background: var(--background-secondary);
			}
		}
	}

	.search-section {
		padding: 1.5rem;
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.search-input-container {
		position: relative;
		margin-bottom: 1.5rem;

		:global(svg) {
			position: absolute;
			left: 1rem;
			top: 50%;
			transform: translateY(-50%);
			color: var(--text-muted);
		}

		input {
			width: 100%;
			padding: 0.75rem 1rem 0.75rem 2.5rem;
			border: 1px solid var(--border);
			border-radius: 6px;
			font-size: 1rem;
			background: var(--background);
			color: var(--text);
			font-family: inherit;

			&:focus {
				outline: none;
				border-color: var(--primary);
			}

			&::placeholder {
				color: var(--text-muted);
			}
		}
	}

	.searching, .no-results, .search-prompt {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		color: var(--text-muted);
		gap: 1rem;

		:global(svg) {
			font-size: 2rem;
			opacity: 0.5;
		}
	}

	.search-results {
		flex: 1;
		overflow-y: auto;
		margin: -0.5rem;
		padding: 0.5rem;
	}

	.movie-result {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.75rem;
		border-radius: 6px;
		cursor: pointer;
		transition: background-color 0.2s ease;

		&:hover {
			background: var(--background-secondary);
		}

		.poster-small {
			width: 50px;
			height: 75px;
			flex-shrink: 0;
		}

		.movie-info {
			flex: 1;
			min-width: 0;

			h4 {
				margin: 0 0 0.25rem 0;
				font-size: 1rem;
				line-height: 1.2;
				color: var(--text);
			}

			.release-year {
				margin: 0 0 0.5rem 0;
				color: var(--text-muted);
				font-size: 0.9rem;
			}

			.overview {
				margin: 0;
				color: var(--text-muted);
				font-size: 0.85rem;
				line-height: 1.3;
			}
		}

		:global(svg) {
			color: var(--text-muted);
			transform: rotate(-90deg);
		}
	}

	.nomination-section {
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.back-btn {
		background: var(--background);
		border: 1px solid var(--border);
		padding: 0.5rem 1rem;
		border-radius: 4px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		align-self: flex-start;
		color: var(--text);
		font-family: inherit;
		font-size: 0.9rem;
		transition: all 0.2s ease;

		&:hover {
			background: var(--background-secondary);
		}

		:global(svg) {
			transform: rotate(90deg);
		}
	}

	.selected-movie {
		display: flex;
		gap: 1rem;

		.poster-large {
			width: 120px;
			height: 180px;
			flex-shrink: 0;
		}

		.movie-details {
			flex: 1;

			h4 {
				margin: 0 0 0.5rem 0;
				font-size: 1.25rem;
				color: var(--text);
			}

			.release-year {
				margin: 0 0 1rem 0;
				color: var(--text-muted);
			}

			.overview {
				margin: 0;
				color: var(--text-muted);
				line-height: 1.4;
				font-size: 0.9rem;
			}
		}
	}

	.reason-section {
		label {
			display: block;
			margin-bottom: 0.5rem;
			font-weight: 500;
			color: var(--text);
		}

		textarea {
			width: 100%;
			min-height: 100px;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 6px;
			font-family: inherit;
			font-size: 0.9rem;
			background: var(--background);
			color: var(--text);
			resize: vertical;

			&:focus {
				outline: none;
				border-color: var(--primary);
			}

			&::placeholder {
				color: var(--text-muted);
			}
		}

		small {
			color: var(--text-muted);
			font-size: 0.8rem;
			margin-top: 0.25rem;
			display: block;
		}
	}

	.action-buttons {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;

		button {
			padding: 0.75rem 1.5rem;
			border-radius: 6px;
			font-size: 1rem;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 0.5rem;
			font-family: inherit;
			transition: all 0.2s ease;
		}

		.cancel-btn {
			background: var(--background);
			border: 1px solid var(--border);
			color: var(--text);

			&:hover {
				background: var(--background-secondary);
			}
		}

		.submit-btn {
			background: var(--primary);
			border: none;
			color: white;

			&:hover {
				background: var(--primary-dark);
			}
		}
	}

	@media (max-width: 768px) {
		.search-modal {
			max-height: 90vh;
		}

		.selected-movie {
			flex-direction: column;
			text-align: center;

			.poster-large {
				align-self: center;
			}
		}

		.action-buttons {
			flex-direction: column;

			button {
				justify-content: center;
			}
		}
	}
</style>