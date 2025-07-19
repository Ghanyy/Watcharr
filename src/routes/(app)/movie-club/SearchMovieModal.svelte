<script lang="ts">
	import { createEventDispatcher } from "svelte";
	import { onMount } from "svelte";
	import type { Content } from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { checkPreviousWinner } from "@/lib/util/api";
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
	let isSelectedMovieExcluded = false; // Track if selected movie is excluded
	let exclusionMessage = ""; // Message explaining why movie is excluded
	let checkingExclusion = false; // Loading state for exclusion check

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
					q: query,
				},
			});
			// Map TMDB response to our Content interface
			const results = response.data.results || [];
			searchResults = results.map((movie: any) => ({
				...movie,
				tmdbId: movie.id, // Ensure tmdbId is set from TMDB's id field
				type: "movie" as const,
			}));
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

	async function selectMovie(movie: Content) {
		selectedMovie = movie;
		reason = "";
		isSelectedMovieExcluded = false;
		exclusionMessage = "";
		checkingExclusion = true;

		try {
			const result = await checkPreviousWinner(movie.tmdbId);
			if (result) {
				isSelectedMovieExcluded = true;
				exclusionMessage = "This movie has already won in a previous cycle or is currently the winner of an active cycle.";
			}
		} catch (err) {
			console.error("Failed to check movie exclusion:", err);
			// On error, allow nomination but let server validation handle it
		} finally {
			checkingExclusion = false;
		}
	}

	function submitNomination() {
		if (!selectedMovie || isSelectedMovieExcluded) return;

		dispatch("movieSelected", {
			content: selectedMovie,
			reason: reason.trim(),
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
						<div
							class="movie-result"
							role="button"
							tabindex="0"
							on:click={() => selectMovie(movie)}
							on:keydown={(e) => {
								if (e.key === "Enter" || e.key === " ") {
									e.preventDefault();
									selectMovie(movie);
								}
							}}
						>
							<div class="poster-small">
								<Poster
									media={movie}
									showRating={false}
									disableInteraction={true}
									small={true}
								/>
							</div>
							<div class="movie-info">
								<div class="movie-title-row">
									<h4>{movie.title}</h4>
								</div>
								<p class="release-year">
									{movie.release_date
										? new Date(movie.release_date).getFullYear()
										: "Unknown"}
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
					<Poster
						media={selectedMovie}
						showRating={false}
						disableInteraction={true}
						small={false}
					/>
				</div>
				<div class="movie-details">
					<h4>{selectedMovie.title}</h4>
					<p class="release-year">
						{selectedMovie.release_date
							? new Date(selectedMovie.release_date).getFullYear()
							: "Unknown"}
					</p>
					{#if selectedMovie.overview}
						<p class="overview">{selectedMovie.overview}</p>
					{/if}
				</div>
			</div>

			<!-- Exclusion Check Status -->
			{#if checkingExclusion}
				<div class="checking-status">
					<Spinner />
					<p>Checking eligibility...</p>
				</div>
			{:else if isSelectedMovieExcluded}
				<div class="exclusion-message">
					<Icon icon="block" />
					<p>{exclusionMessage}</p>
				</div>
			{/if}

			<div class="reason-section">
				<label for="reason">Why are you nominating this movie? (Optional)</label
				>
				<textarea
					id="reason"
					placeholder="Share why you think this would be a great choice..."
					bind:value={reason}
					maxlength="500"
					disabled={isSelectedMovieExcluded}
				></textarea>
				<small>{reason.length}/500 characters</small>
			</div>

			<div class="action-buttons">
				<button class="cancel-btn" on:click={() => dispatch("close")}>
					Cancel
				</button>
				<button 
					class="submit-btn" 
					on:click={submitNomination}
					disabled={isSelectedMovieExcluded || checkingExclusion}
				>
					<Icon icon="add" />
					{#if checkingExclusion}
						Checking...
					{:else if isSelectedMovieExcluded}
						Cannot Nominate
					{:else}
						Nominate This Movie
					{/if}
				</button>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.search-modal {
		width: 100%;
		height: 100%;
		min-height: 500px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		background: var(--background);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		border: 1px solid var(--border);

		// CSS custom properties for consistent design system
		--space-xs: 0.25rem;
		--space-sm: 0.5rem;
		--space-md: 1rem;
		--space-lg: 1.5rem;
		--space-xl: 2rem;
		--space-2xl: 3rem;

		--radius-sm: 4px;
		--radius-md: 8px;
		--radius-lg: 12px;
		--radius-xl: 16px;
		--radius-full: 50%;

		--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
		--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
			0 2px 4px -1px rgba(0, 0, 0, 0.06);
		--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
			0 4px 6px -2px rgba(0, 0, 0, 0.05);

		@media (prefers-color-scheme: dark) {
			--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
			--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.4),
				0 2px 4px -1px rgba(0, 0, 0, 0.3);
			--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.5),
				0 4px 6px -2px rgba(0, 0, 0, 0.4);
		}
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-lg);
		border-bottom: 1px solid var(--border);
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--primary);
			opacity: 0.01;
			pointer-events: none;
		}

		h3 {
			margin: 0;
			color: var(--text);
			font-weight: 700;
			font-size: 1.25rem;
		}

		.close-btn {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 32px;
			height: 32px;
			background: none;
			border: none;
			cursor: pointer;
			color: var(--text-muted);
			border-radius: var(--radius-md);
			transition: all 0.2s ease;
			border: 1px solid transparent;

			&:hover {
				color: var(--text);
				background: var(--background);
				border-color: var(--border);
				box-shadow: var(--shadow-sm);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 2px var(--primary);
			}
		}
	}

	.search-section {
		padding: var(--space-lg);
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.search-input-container {
		position: relative;
		margin-bottom: 0;

		:global(svg) {
			position: absolute;
			left: var(--space-md);
			top: 50%;
			transform: translateY(-50%);
			color: var(--text-muted);
			z-index: 1;
			font-size: 1.1rem;
		}

		input {
			width: 100%;
			padding: var(--space-md) var(--space-md) var(--space-md) 2.75rem;
			border: 2px solid var(--border);
			border-radius: var(--radius-lg);
			font-size: 1rem;
			background: var(--background);
			color: var(--text);
			font-family: inherit;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);

			&:focus {
				outline: none;
				border-color: var(--primary);
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(59, 130, 246, 0.1);
				background: var(--background);
			}

			&::placeholder {
				color: var(--text-muted);
				font-weight: 400;
			}

			&:hover:not(:focus) {
				box-shadow: var(--shadow-md);
			}
		}
	}

	.searching,
	.no-results,
	.search-prompt {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		color: var(--text-muted);
		gap: var(--space-md);
		padding: var(--space-xl) var(--space-md);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);
		position: relative;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: var(--text-muted);
			opacity: 0.02;
			border-radius: inherit;
			pointer-events: none;
		}

		:global(svg) {
			font-size: 2.5rem;
			opacity: 0.4;
			color: var(--text-muted);
		}

		p {
			margin: 0;
			text-align: center;
			line-height: 1.4;
			font-size: 0.95rem;
		}
	}

	.search-results {
		flex: 1;
		overflow-y: auto;
		margin: 0;
		padding: var(--space-sm);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		min-height: 300px;
		max-height: none;
		box-shadow: var(--shadow-sm);
	}

	.movie-result {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		cursor: pointer;
		transition: all 0.3s ease;
		border: 1px solid var(--border);
		background: var(--background);
		margin-bottom: var(--space-sm);
		box-shadow: var(--shadow-sm);
		position: relative;
		overflow: hidden;

		&:last-child {
			margin-bottom: 0;
		}

		&:hover:not(.ineligible) {
			background: var(--background);
			transform: translateY(-2px);
			box-shadow: var(--shadow-lg);

			&::before {
				content: "";
				position: absolute;
				inset: 0;
				background: var(--primary);
				opacity: 0.02;
				border-radius: inherit;
				pointer-events: none;
			}
		}

		&.ineligible {
			opacity: 0.6;
			cursor: not-allowed;
			filter: grayscale(0.3);
			border-color: var(--border);

			&::after {
				content: "";
				position: absolute;
				inset: 0;
				background: repeating-linear-gradient(
					45deg,
					transparent,
					transparent 10px,
					rgba(255, 0, 0, 0.1) 10px,
					rgba(255, 0, 0, 0.1) 20px
				);
				pointer-events: none;
				border-radius: inherit;
			}
		}

		.poster-small {
			width: 60px;
			height: 90px;
			flex-shrink: 0;
			border-radius: 6px;
			overflow: hidden;
			position: relative;
			border: 1px solid var(--border);
			background: var(--background);
			transition: border-color 0.3s ease;

			&:hover {
				border-color: var(--primary);
			}

			:global(li) {
				list-style: none;
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
				cursor: default !important;
			}

			:global(.container) {
				width: 100% !important;
				height: 100% !important;
				min-width: unset !important;
				transform: none !important;
				position: static !important;
				border-radius: 0 !important;
				transition: none !important;
			}

			:global(.active .container) {
				transform: none !important;
			}

			:global(.inner) {
				opacity: 0 !important;
				pointer-events: none !important;
			}

			:global(img) {
				width: 100%;
				height: 100%;
				object-fit: cover;
				object-position: center;
				border-radius: 0;
				filter: none !important;
				mix-blend-mode: normal !important;
			}
		}

		.movie-info {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
			gap: var(--space-xs);
			position: relative;
			z-index: 1;

			.movie-title-row {
				display: flex;
				align-items: flex-start;
				gap: var(--space-sm);
				flex-wrap: wrap;
				margin-bottom: var(--space-xs);
			}

			h4 {
				margin: 0;
				font-size: 1rem;
				line-height: 1.2;
				color: var(--text);
				font-weight: 600;
				word-wrap: break-word;
				transition: color 0.2s ease;
				flex: 1;
				min-width: 0;
			}

			.previous-winner-badge {
				display: inline-flex;
				align-items: center;
				gap: 2px;
				padding: 2px var(--space-xs);
				background: linear-gradient(135deg, #fbbf24, #f59e0b);
				color: white;
				font-size: 0.7rem;
				font-weight: 600;
				border-radius: var(--radius-sm);
				white-space: nowrap;
				flex-shrink: 0;
				box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
				text-shadow: 0 1px 1px rgba(0, 0, 0, 0.2);

				:global(svg) {
					font-size: 0.7rem;
					color: white;
				}
			}

			.ineligible-message {
				margin: var(--space-xs) 0 0 0;
				color: var(--danger, #dc3545);
				font-size: 0.75rem;
				font-weight: 500;
				font-style: italic;
				line-height: 1.3;
			}

			.release-year {
				margin: 0;
				color: var(--text-muted);
				font-size: 0.875rem;
				font-weight: 500;
				padding: var(--space-xs) var(--space-sm);
				border-radius: var(--radius-md);
				border: 1px solid var(--border);
				width: fit-content;
			}

			.overview {
				margin: 0;
				color: var(--text-muted);
				font-size: 0.85rem;
				line-height: 1.4;
				display: -webkit-box;
				-webkit-line-clamp: 2;
				-webkit-box-orient: vertical;
				overflow: hidden;
				text-overflow: ellipsis;
				margin-top: var(--space-xs);
			}
		}

		:global(svg) {
			color: var(--text-muted);
			transform: rotate(-90deg);
			transition: all 0.2s ease;
			position: relative;
			z-index: 1;
		}

		&:hover:not(.ineligible) :global(svg) {
			color: var(--primary);
			transform: rotate(-90deg) scale(1.1);
		}
	}

	.nomination-section {
		padding: var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}

	.back-btn {
		display: inline-flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-sm) var(--space-md);
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		color: var(--text);
		font-family: inherit;
		font-size: 0.9rem;
		font-weight: 500;
		cursor: pointer;
		align-self: flex-start;
		transition: all 0.2s ease;
		box-shadow: var(--shadow-sm);

		&:hover {
			background: var(--background);
			color: var(--primary);
			transform: translateX(-2px);
			box-shadow: var(--shadow-md);
		}

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-md),
				0 0 0 2px var(--primary);
		}

		:global(svg) {
			transform: rotate(90deg);
			transition: transform 0.2s ease;
		}

		&:hover :global(svg) {
			transform: rotate(90deg) translateX(-1px);
		}
	}

	.selected-movie {
		display: flex;
		gap: var(--space-lg);
		padding: var(--space-md);
		background: var(--background);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		box-shadow: var(--shadow-md);
		transition: all 0.2s ease;

		&:hover {
			box-shadow: var(--shadow-lg);
		}

		.poster-large {
			width: 120px;
			height: 180px;
			flex-shrink: 0;
			position: relative;
			border-radius: 8px;
			overflow: hidden;
			border: 1px solid var(--border);
			background: var(--background);
			transition: border-color 0.3s ease;

			&:hover {
				border-color: var(--primary);
			}

			:global(li) {
				list-style: none;
				margin: 0;
				padding: 0;
				width: 100%;
				height: 100%;
				cursor: default !important;
			}

			:global(.container) {
				width: 100% !important;
				height: 100% !important;
				min-width: unset !important;
				transform: none !important;
				position: static !important;
				border-radius: 0 !important;
				transition: none !important;
			}

			:global(.active .container) {
				transform: none !important;
			}

			:global(.inner) {
				opacity: 0 !important;
				pointer-events: none !important;
			}

			:global(img) {
				width: 100%;
				height: 100%;
				object-fit: cover;
				object-position: center;
				border-radius: 0;
				filter: none !important;
				mix-blend-mode: normal !important;
			}
		}

		.movie-details {
			flex: 1;
			display: flex;
			flex-direction: column;
			gap: var(--space-md);

			h4 {
				margin: 0;
				font-size: 1.25rem;
				color: var(--text);
				font-weight: 700;
				line-height: 1.2;
			}

			.release-year {
				margin: 0;
				color: var(--text-muted);
				font-size: 1rem;
				font-weight: 500;
				padding: var(--space-xs) var(--space-sm);
				border-radius: var(--radius-md);
				border: 1px solid var(--border);
				width: fit-content;
			}

			.overview {
				margin: 0;
				color: var(--text-muted);
				line-height: 1.5;
				font-size: 0.9rem;
				max-height: 120px;
				overflow-y: auto;
				padding: var(--space-sm);
				border-radius: var(--radius-md);
				border: 1px solid var(--border);
				border-left: 3px solid var(--primary);
				position: relative;

				&::before {
					content: "";
					position: absolute;
					inset: 0;
					background: var(--primary);
					opacity: 0.02;
					border-radius: inherit;
					pointer-events: none;
				}
			}
		}
	}

	.reason-section {
		background: var(--background);
		padding: var(--space-md);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);

		label {
			display: block;
			margin-bottom: var(--space-sm);
			font-weight: 600;
			color: var(--text);
			font-size: 0.95rem;
		}

		textarea {
			width: 100%;
			min-height: 100px;
			padding: var(--space-sm);
			border: 2px solid var(--border);
			border-radius: var(--radius-md);
			font-family: inherit;
			font-size: 0.9rem;
			background: var(--background);
			color: var(--text);
			resize: vertical;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);

			&:focus {
				outline: none;
				border-color: var(--primary);
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(59, 130, 246, 0.1);
				background: var(--background);
			}

			&::placeholder {
				color: var(--text-muted);
				font-style: italic;
			}

			&:hover:not(:focus) {
				box-shadow: var(--shadow-md);
			}
		}

		small {
			color: var(--text-muted);
			font-size: 0.8rem;
			margin-top: var(--space-xs);
			display: block;
			font-weight: 500;
		}
	}

	.action-buttons {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
		padding: var(--space-md);
		background: var(--background);
		border-top: 1px solid var(--border);
		border-radius: 0 0 var(--radius-lg) var(--radius-lg);
		margin: var(--space-md) -#{var(--space-lg)} -#{var(--space-lg)} -#{var(
				--space-lg
			)};

		button {
			display: inline-flex;
			align-items: center;
			gap: var(--space-sm);
			padding: var(--space-sm) var(--space-lg);
			border-radius: var(--radius-md);
			font-size: 1rem;
			font-weight: 500;
			cursor: pointer;
			font-family: inherit;
			transition: all 0.2s ease;
			box-shadow: var(--shadow-sm);
			min-width: 120px;
			justify-content: center;
		}

		.cancel-btn {
			background: var(--background);
			border: 1px solid var(--border);
			color: var(--text);

			&:hover {
				background: var(--background);
				border-color: var(--text-muted);
				transform: translateY(-1px);
				box-shadow: var(--shadow-md);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 2px var(--text-muted);
			}
		}

		.submit-btn {
			background: var(--primary);
			border: none;
			color: white;

			&:hover {
				background: var(--primary-dark, var(--primary));
				transform: translateY(-1px);
				box-shadow: var(--shadow-md);
			}

			&:focus {
				outline: none;
				box-shadow:
					var(--shadow-md),
					0 0 0 3px rgba(59, 130, 246, 0.2);
			}

			&:active {
				transform: translateY(0);
				box-shadow: var(--shadow-sm);
			}
		}
	}

	@media (max-width: 768px) {
		.search-modal {
			min-height: 400px;
		}

		.modal-header {
			padding: var(--space-md);

			h3 {
				font-size: 1.125rem;
			}

			.close-btn {
				width: 28px;
				height: 28px;
			}
		}

		.search-section {
			padding: var(--space-md);
			gap: var(--space-sm);
		}

		.nomination-section {
			padding: var(--space-md);
			gap: var(--space-md);
		}

		.search-input-container {
			:global(svg) {
				left: var(--space-sm);
				font-size: 1rem;
			}

			input {
				padding: var(--space-sm) var(--space-sm) var(--space-sm) 2.5rem;
				font-size: 0.95rem;
			}
		}

		.searching,
		.no-results,
		.search-prompt {
			padding: var(--space-lg) var(--space-sm);

			:global(svg) {
				font-size: 2rem;
			}

			p {
				font-size: 0.9rem;
			}
		}

		.search-results {
			max-height: 300px;
			padding: var(--space-xs);
		}

		.movie-result {
			padding: var(--space-sm);
			gap: var(--space-sm);

			.poster-small {
				width: 50px;
				height: 75px;
			}

			.movie-info {
				gap: 2px;

				h4 {
					font-size: 0.9rem;
				}

				.release-year {
					font-size: 0.8rem;
					padding: 2px var(--space-xs);
				}

				.overview {
					font-size: 0.75rem;
					-webkit-line-clamp: 2;
				}
			}
		}

		.selected-movie {
			flex-direction: column;
			text-align: center;
			gap: var(--space-md);
			padding: var(--space-sm);

			.poster-large {
				align-self: center;
				width: 100px;
				height: 150px;
			}

			.movie-details {
				text-align: center;
				gap: var(--space-sm);

				h4 {
					font-size: 1.125rem;
				}

				.release-year {
					font-size: 0.9rem;
					align-self: center;
				}

				.overview {
					font-size: 0.85rem;
					max-height: 100px;
				}
			}
		}

		.reason-section {
			padding: var(--space-sm);

			label {
				font-size: 0.9rem;
			}

			textarea {
				min-height: 80px;
				font-size: 0.85rem;
			}

			small {
				font-size: 0.75rem;
			}
		}

		.action-buttons {
			flex-direction: column;
			gap: var(--space-sm);
			padding: var(--space-sm);
			margin: var(--space-sm) -#{var(--space-md)} -#{var(--space-md)} -#{var(
					--space-md
				)};

			button {
				justify-content: center;
				padding: var(--space-sm) var(--space-md);
				font-size: 0.9rem;
				min-width: unset;
			}
		}
	}

	.checking-status {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		background: var(--background-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		color: var(--text-muted);
		margin-bottom: var(--space-md);

		:global(.spinner) {
			font-size: 1.2rem;
		}

		p {
			margin: 0;
			font-size: 0.9rem;
		}
	}

	.exclusion-message {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: var(--radius-md);
		color: #dc2626;
		margin-bottom: var(--space-md);

		@media (prefers-color-scheme: dark) {
			background: #450a0a;
			border-color: #7f1d1d;
			color: #fca5a5;
		}

		:global(svg) {
			font-size: 1.2rem;
			flex-shrink: 0;
		}

		p {
			margin: 0;
			font-size: 0.9rem;
			font-weight: 500;
			line-height: 1.4;
		}
	}

	.reason-section {
		textarea:disabled {
			opacity: 0.5;
			cursor: not-allowed;
			background: var(--background-secondary);
		}
	}

	.submit-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		background: var(--text-muted);
		
		&:hover {
			background: var(--text-muted);
			transform: none;
			box-shadow: var(--shadow-sm);
		}
	}
</style>
