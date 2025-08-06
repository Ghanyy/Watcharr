<script lang="ts">
	import type { MovieClubCycleResponse } from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";

	export let cycleData: MovieClubCycleResponse;

	$: results = cycleData.voteResults || [];
	$: winner = results.length > 0 ? results[0] : null;
	$: hasVotes = results.some((r) => r.totalVotes > 0);
	$: cycleRatings = cycleData.cycleRatings || [];
	$: validRatings = cycleRatings.filter((rating) => rating.rating > 0);
	$: thoughtsOnlyEntries = cycleRatings.filter(
		(rating) => rating.rating <= 0 && rating.thoughts && rating.thoughts.trim(),
	);
	$: hasCycleRatings = cycleRatings.length > 0;
	$: shouldShowClubAverage = validRatings.length >= 2;
	$: clubAverageRating = shouldShowClubAverage
		? (
				validRatings.reduce((sum, rating) => sum + rating.rating, 0) /
				validRatings.length
			).toFixed(1)
		: null;
</script>

<div class="results-section">
	<div class="section-header">
		<h3>
			<Icon icon="play" />
			{#if cycleData.cycle.isAdHoc}
				Ad-hoc Session
			{:else}
				Voting Results
			{/if}
		</h3>
		<p>
			{#if cycleData.cycle.isAdHoc}
				This movie was directly selected for an ad-hoc session!
			{:else}
				Here are the results from this cycle's voting!
			{/if}
		</p>
	</div>

	{#if cycleData.cycle.isAdHoc && cycleData.cycle.winnerContent}
		<!-- Ad-hoc Chosen Movie Section -->
		<div class="chosen-movie-section">
			<div class="chosen-badge">
				<Icon icon="sparkles" />
				<span>Chosen Movie</span>
			</div>

			<div class="chosen-card">
				<div class="chosen-poster">
					<Poster media={cycleData.cycle.winnerContent} showRating={false} />
				</div>
				<div class="chosen-info">
					<h4>{cycleData.cycle.winnerContent.title}</h4>
					<p class="chosen-year">
						{cycleData.cycle.winnerContent.release_date
							? new Date(cycleData.cycle.winnerContent.release_date).getFullYear()
							: "Unknown"}
					</p>
					{#if cycleData.cycle.winnerContent.overview}
						<p class="chosen-overview">
							{cycleData.cycle.winnerContent.overview}
						</p>
					{/if}
				</div>
			</div>
		</div>
	{:else if !hasVotes}
		<div class="no-votes">
			<Icon icon="check" />
			<p>No votes were cast this cycle.</p>
		</div>
	{:else}
		<!-- Winner Section -->
		{#if winner}
			<div class="winner-section">
				<div class="winner-badge">
					<Icon icon="sparkles" />
					<span>Winner</span>
				</div>

				<div class="winner-card">
					<div class="winner-poster">
						<Poster
							media={winner.content}
							showRating={false}
							disableInteraction={true}
							fluidSize={true}
						/>
					</div>

					<div class="winner-details">
						<h2>{winner.content.title}</h2>
						<div class="winner-stats">
							<div class="stat">
								<span class="stat-value">{winner.totalVotes}</span>
								<span class="stat-label">Total Votes</span>
							</div>
							<div class="stat">
								<span class="stat-value">{winner.weightedScore.toFixed(1)}</span
								>
								<span class="stat-label">Score</span>
							</div>
							<div class="stat">
								<span class="stat-value">{winner.firstChoice}</span>
								<span class="stat-label">1st Choice</span>
							</div>
						</div>

						{#if winner.content.overview}
							<p class="winner-overview">{winner.content.overview}</p>
						{/if}
					</div>
				</div>
			</div>
		{/if}

		<!-- Cycle Ratings Section -->
		{#if hasCycleRatings}
			<div class="cycle-ratings-section">
				{#if shouldShowClubAverage}
					<div class="club-average-rating">
						<span class="average-label">Club rate:</span>
						<span class="average-value"
							><span class="golden-asterisk">*</span
							>{clubAverageRating}/10</span
						>
					</div>
				{/if}

				<details class="cycle-ratings-list">
					<summary class="ratings-header">
						<Icon icon="star" />
						<span
							>Member Ratings ({validRatings.length}{thoughtsOnlyEntries.length >
							0
								? ` + ${thoughtsOnlyEntries.length} thoughts`
								: ""})</span
						>
						<Icon icon="chevron" />
					</summary>

					<div class="ratings-content">
						{#each cycleRatings as rating}
							<div class="rating-item">
								<div class="rating-user">
									{#if rating.user?.id && rating.user?.username}
										<a
											href="/lists/{rating.user.id}/{rating.user.username}"
											class="username">{rating.user.username}</a
										>
									{:else}
										<span class="username">Unknown User</span>
									{/if}
									{#if rating.rating > 0}
										<span class="rating-score">{rating.rating}/10</span>
									{:else}
										<span class="thoughts-only">Thoughts only</span>
									{/if}
								</div>
								{#if rating.thoughts && rating.thoughts.trim()}
									<details class="thoughts-details">
										<summary class="thoughts-toggle">
											<Icon icon="document" />
											<span>View thoughts</span>
										</summary>
										<div class="thoughts-content">
											<p>{rating.thoughts}</p>
										</div>
									</details>
								{/if}
							</div>
						{/each}
					</div>
				</details>
			</div>
		{/if}

		<!-- Full Results -->
		<div class="full-results">
			<h4>Complete Results</h4>
			<div class="results-table">
				<div class="results-header">
					<span class="rank">Rank</span>
					<span class="movie">Movie</span>
					<span class="votes">Votes</span>
					<span class="score">Score</span>
					<span class="breakdown">Vote Breakdown</span>
				</div>

				{#each results as result, index}
					<div class="result-row" class:winner={index === 0}>
						<div class="rank">
							#{index + 1}
						</div>

						<div class="movie">
							<div class="movie-info">
								<h5>
									<a href="/movie/{result.content.tmdbId}" class="movie-link">
										{result.content.title}
									</a>
								</h5>
								<span class="release-year">
									{result.content.release_date
										? new Date(result.content.release_date).getFullYear()
										: "Unknown"}
								</span>
							</div>
						</div>

						<div class="votes">
							{result.totalVotes}
						</div>

						<div class="score">
							{result.weightedScore.toFixed(1)}
						</div>

						<div class="breakdown">
							<div class="vote-breakdown">
								{#if result.firstChoice > 0}
									<span class="choice first">
										{result.firstChoice} × 1st
									</span>
								{/if}
								{#if result.secondChoice > 0}
									<span class="choice second">
										{result.secondChoice} × 2nd
									</span>
								{/if}
								{#if result.thirdChoice > 0}
									<span class="choice third">
										{result.thirdChoice} × 3rd
									</span>
								{/if}
								{#if result.totalVotes === 0}
									<span class="no-votes">No votes</span>
								{/if}
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Voting Summary -->
		<div class="voting-summary">
			<h4>Voting Summary</h4>
			<div class="summary-stats">
				<div class="summary-stat">
					<span class="stat-value"
						>{results.reduce((sum, r) => sum + r.totalVotes, 0)}</span
					>
					<span class="stat-label">Total Votes Cast</span>
				</div>
				<div class="summary-stat">
					<span class="stat-value">{results.length}</span>
					<span class="stat-label">Movies Nominated</span>
				</div>
				<div class="summary-stat">
					<span class="stat-value"
						>{results.filter((r) => r.totalVotes > 0).length}</span
					>
					<span class="stat-label">Movies with Votes</span>
				</div>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.results-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);

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

	.section-header {
		text-align: center;
		margin-bottom: var(--space-lg);

		h3 {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: var(--space-sm);
			margin: 0 0 var(--space-sm) 0;
			font-size: 1.5rem;
			font-weight: 700;
			color: var(--text);
			line-height: 1.2;
		}

		p {
			color: var(--text-muted);
			margin: 0;
			line-height: 1.4;
			font-size: 0.95rem;
		}
	}

	.no-votes {
		text-align: center;
		padding: var(--space-2xl) var(--space-md);
		color: var(--text-muted);
		background: var(--background-secondary);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);

		:global(svg) {
			font-size: 3rem;
			margin-bottom: var(--space-md);
			opacity: 0.4;
			color: var(--text-muted);
		}

		p {
			margin: 0;
			line-height: 1.4;
		}
	}

	.winner-section {
		position: relative;
		margin-bottom: var(--space-xl);
	}

	.winner-badge {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		background: linear-gradient(135deg, #ffd700, #ffed4e);
		color: #000;
		padding: var(--space-sm) var(--space-lg);
		border-radius: 20px;
		font-weight: 700;
		font-size: 1.1rem;
		width: fit-content;
		margin: 0 auto var(--space-lg) auto;
		box-shadow: var(--shadow-lg);
		border: 2px solid #ffd700;
		text-transform: uppercase;
		letter-spacing: 0.025em;
		transition: all 0.3s ease;

		&:hover {
			transform: scale(1.05);
			box-shadow: 0 8px 25px rgba(255, 215, 0, 0.4);
		}

		:global(svg) {
			font-size: 1.25rem;
		}
	}

	.winner-card {
		background: var(--background);
		border: 2px solid #ffd700;
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		display: flex;
		gap: var(--space-xl);
		align-items: center;
		box-shadow: var(--shadow-lg);
		position: relative;
		transition: all 0.3s ease;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: linear-gradient(
				135deg,
				rgba(255, 215, 0, 0.1),
				rgba(255, 237, 78, 0.05)
			);
			border-radius: inherit;
			pointer-events: none;
		}

		&:hover {
			transform: translateY(-2px);
			box-shadow: 0 12px 32px rgba(255, 215, 0, 0.3);
		}
	}

	.winner-poster {
		position: relative;
		width: 200px;
		height: 300px;
		flex-shrink: 0;
		z-index: 1;
		border-radius: var(--radius-lg);
		overflow: hidden;
		box-shadow: var(--shadow-lg);

		:global(li) {
			list-style: none;
			margin: 0;
			padding: 0;
			width: 100%;
			height: 100%;
		}

		:global(.container) {
			width: 100% !important;
			height: 100% !important;
			min-width: unset !important;
		}

		// Disable the zoom/scale effect on hover for winner poster
		:global(.active .container) {
			transform: none !important;
			z-index: 1 !important;
		}
	}

	.cycle-ratings-section {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);
		border-left: 4px solid var(--warning, #f59e0b);
	}

	.club-average-rating {
		background: var(--background);
		border: 1px solid var(--warning, #f59e0b);
		border-radius: var(--radius-lg);
		padding: var(--space-md);
		margin-bottom: var(--space-md);
		text-align: center;
		box-shadow: var(--shadow-sm);

		.average-label {
			font-weight: 600;
			color: var(--text);
			margin-right: var(--space-sm);
			font-size: 1.1rem;
		}

		.average-value {
			font-weight: 700;
			font-size: 1.5rem;
			color: var(--warning, #f59e0b);
			text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);

			.golden-asterisk {
				font-family: Rampart One;
				-webkit-text-stroke: 1px var(--warning, #f59e0b);
				font-size: 40px;
				line-height: 0.7;
				margin-right: 0.1em;
				color: var(--warning, #f59e0b);
				vertical-align: baseline;
				transform: translateY(0.4em);
			}
		}
	}

	.cycle-ratings-list {
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		background: var(--background);
		overflow: hidden;
		transition: all 0.2s ease;

		&:hover {
			box-shadow: var(--shadow-md);
		}

		&[open] {
			.ratings-header {
				border-bottom: 1px solid var(--border);

				:global(.chevron) {
					transform: rotate(180deg);
				}
			}
		}
	}

	.ratings-header {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		background: var(--background);
		cursor: pointer;
		font-weight: 600;
		color: var(--text);
		border: none;
		transition: all 0.2s ease;
		list-style: none;

		&:hover {
			background: var(--background-secondary);
		}

		:global(svg) {
			color: var(--warning, #f59e0b);
			transition: transform 0.2s ease;
		}

		:global(.chevron) {
			margin-left: auto;
			font-size: 0.875rem;
		}

		span {
			font-size: 1rem;
		}
	}

	.ratings-content {
		padding: var(--space-md);
		background: var(--background);
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.rating-item {
		padding: var(--space-md);
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		transition: all 0.2s ease;

		&:hover {
			transform: translateY(-1px);
			box-shadow: var(--shadow-sm);
			background: var(--background-secondary);
		}
	}

	.rating-user {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-sm);

		.username {
			font-weight: 600;
			color: var(--text);
			font-size: 0.95rem;
			text-decoration: none;
			transition: all 0.2s ease;
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-sm);
			margin: -var(--space-xs) -var(--space-sm);

			&:hover {
				color: var(--primary);
				background: var(--background-secondary);
				text-decoration: none;
				transform: translateX(1px);
			}

			&:focus {
				outline: none;
				box-shadow: 0 0 0 2px var(--primary);
			}
		}

		.rating-score {
			font-weight: 700;
			color: var(--warning, #f59e0b);
			background: var(--background-secondary);
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-md);
			border: 1px solid var(--warning, #f59e0b);
			font-size: 0.875rem;
		}

		.thoughts-only {
			font-weight: 600;
			color: var(--text-muted);
			background: var(--background-secondary);
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-md);
			border: 1px solid var(--border);
			font-size: 0.875rem;
			font-style: italic;
		}
	}

	.thoughts-details {
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
		background: var(--background);
		overflow: hidden;
		transition: all 0.2s ease;

		&[open] {
			.thoughts-toggle {
				border-bottom: 1px solid var(--border);
			}
		}
	}

	.thoughts-toggle {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-sm);
		background: var(--background);
		cursor: pointer;
		font-size: 0.875rem;
		color: var(--text-muted);
		border: none;
		transition: all 0.2s ease;
		list-style: none;

		&:hover {
			background: var(--background-secondary);
			color: var(--text);
		}

		:global(svg) {
			font-size: 0.875rem;
			color: var(--text-muted);
		}
	}

	.thoughts-content {
		padding: var(--space-sm);
		background: var(--background);

		p {
			margin: 0;
			line-height: 1.5;
			color: var(--text);
			font-size: 0.9rem;
			white-space: pre-wrap;
		}
	}

	.winner-details {
		flex: 1;
		position: relative;
		z-index: 1;

		h2 {
			margin: 0 0 var(--space-md) 0;
			font-size: 2rem;
			color: var(--text);
			font-weight: 800;
			line-height: 1.1;
		}

		.winner-stats {
			display: flex;
			gap: var(--space-xl);
			margin-bottom: var(--space-lg);
			flex-wrap: wrap;

			.stat {
				text-align: center;
				padding: var(--space-sm) var(--space-md);
				background: var(--background-secondary);
				border-radius: var(--radius-lg);
				border: 1px solid var(--border);
				box-shadow: var(--shadow-sm);
				transition: all 0.2s ease;
				min-width: 80px;

				&:hover {
					transform: translateY(-2px);
					box-shadow: var(--shadow-md);
				}

				.stat-value {
					display: block;
					font-size: 1.5rem;
					font-weight: 700;
					color: #ffd700;
					text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
					margin-bottom: var(--space-xs);
				}

				.stat-label {
					font-size: 0.875rem;
					color: var(--text-muted);
					font-weight: 500;
					text-transform: uppercase;
					letter-spacing: 0.025em;
				}
			}
		}

		.winner-overview {
			color: var(--text-muted);
			line-height: 1.5;
			margin: 0;
			padding: var(--space-md);
			background: var(--background-secondary);
			border-radius: var(--radius-md);
			border-left: 3px solid #ffd700;
			font-style: italic;
		}
	}

	.full-results {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);

		h4 {
			margin: 0 0 var(--space-md) 0;
			padding-bottom: var(--space-sm);
			border-bottom: 1px solid var(--border);
			font-weight: 600;
			color: var(--text);
			font-size: 1.125rem;
		}
	}

	.results-table {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	.results-header {
		display: grid;
		grid-template-columns: 60px 1fr 80px 80px 200px;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--background-secondary);
		border-radius: var(--radius-md);
		font-weight: 600;
		font-size: 0.875rem;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.025em;
		border: 1px solid var(--border);
	}

	.result-row {
		display: grid;
		grid-template-columns: 60px 1fr 80px 80px 200px;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		align-items: center;
		transition: all 0.2s ease;
		position: relative;

		&:hover {
			background: var(--background-secondary);
			transform: translateY(-1px);
			box-shadow: var(--shadow-md);
		}

		&.winner {
			background: var(--background);
			border-color: #ffd700;
			border-width: 2px;
			box-shadow: var(--shadow-md);

			&::before {
				content: "";
				position: absolute;
				inset: 0;
				background: linear-gradient(
					135deg,
					rgba(255, 215, 0, 0.1),
					rgba(255, 237, 78, 0.05)
				);
				border-radius: inherit;
				pointer-events: none;
			}

			&:hover {
				transform: translateY(-2px);
				box-shadow: var(--shadow-lg);
			}
		}

		.rank {
			text-align: center;
			font-weight: 700;
			color: var(--primary);
			position: relative;
			z-index: 1;
			font-size: 1rem;

			:global(svg) {
				color: #ffd700;
				font-size: 1.2rem;
			}
		}

		.movie {
			display: flex;
			align-items: center;
			position: relative;
			z-index: 1;

			.movie-info {
				h5 {
					margin: 0 0 var(--space-xs) 0;
					font-size: 0.95rem;
					line-height: 1.2;
					font-weight: 600;
				}

				.movie-link {
					color: var(--text);
					text-decoration: none;
					transition: all 0.2s ease;
					font-weight: 600;
					padding: var(--space-xs) var(--space-sm);
					border-radius: var(--radius-md);
					display: inline-block;
					margin: -#{var(--space-xs)} -#{var(--space-sm)};

					&:hover {
						color: var(--primary);
						background: var(--background-secondary);
						text-decoration: none;
						transform: translateX(2px);
					}

					&:focus {
						outline: none;
						box-shadow: 0 0 0 2px var(--primary);
					}
				}

				.release-year {
					font-size: 0.8rem;
					color: var(--text-muted);
					font-weight: 500;
				}
			}
		}

		.votes,
		.score {
			text-align: center;
			font-weight: 600;
			position: relative;
			z-index: 1;
			font-size: 0.95rem;
			color: var(--text);
		}

		.breakdown {
			position: relative;
			z-index: 1;

			.vote-breakdown {
				display: flex;
				flex-wrap: wrap;
				gap: var(--space-xs);

				.choice {
					padding: var(--space-xs) var(--space-sm);
					border-radius: var(--radius-md);
					font-size: 0.75rem;
					font-weight: 600;
					text-transform: uppercase;
					letter-spacing: 0.025em;
					border: 1px solid transparent;
					transition: all 0.2s ease;

					&.first {
						background: var(--success, #22c55e);
						color: white;
						box-shadow: var(--shadow-sm);
					}

					&.second {
						background: var(--warning, #f59e0b);
						color: white;
						box-shadow: var(--shadow-sm);
					}

					&.third {
						background: var(--danger, #ef4444);
						color: white;
						box-shadow: var(--shadow-sm);
					}

					&:hover {
						transform: translateY(-1px);
						box-shadow: var(--shadow-md);
					}
				}

				.no-votes {
					color: var(--text-muted);
					font-size: 0.75rem;
					font-style: italic;
					padding: var(--space-xs) var(--space-sm);
					background: var(--background-secondary);
					border-radius: var(--radius-md);
					border: 1px solid var(--border);
				}
			}
		}
	}

	.chosen-movie-section {
		position: relative;
		margin-bottom: var(--space-xl);
	}

	.chosen-badge {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		background: linear-gradient(135deg, var(--accent, #8b5cf6), #7c3aed);
		color: white;
		padding: var(--space-sm) var(--space-lg);
		border-radius: 20px;
		font-weight: 700;
		font-size: 1.1rem;
		width: fit-content;
		margin: 0 auto var(--space-lg) auto;
		box-shadow: var(--shadow-lg);
		border: 2px solid var(--accent, #8b5cf6);
		text-transform: uppercase;
		letter-spacing: 0.025em;
		transition: all 0.3s ease;

		&:hover {
			transform: scale(1.05);
			box-shadow: 0 8px 25px rgba(139, 92, 246, 0.4);
		}

		:global(svg) {
			font-size: 1.25rem;
		}
	}

	.chosen-card {
		background: var(--background);
		border: 2px solid var(--accent, #8b5cf6);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		display: flex;
		gap: var(--space-xl);
		align-items: center;
		box-shadow: var(--shadow-lg);
		position: relative;
		transition: all 0.3s ease;

		&::before {
			content: "";
			position: absolute;
			inset: 0;
			background: linear-gradient(
				135deg,
				rgba(139, 92, 246, 0.1),
				rgba(124, 58, 237, 0.05)
			);
			border-radius: inherit;
			pointer-events: none;
		}

		&:hover {
			transform: translateY(-2px);
			box-shadow: 0 12px 32px rgba(139, 92, 246, 0.3);
		}
	}

	.chosen-poster {
		position: relative;
		width: 200px;
		height: 300px;
		flex-shrink: 0;
		z-index: 1;
		border-radius: var(--radius-lg);
		overflow: hidden;
		box-shadow: var(--shadow-lg);

		:global(li) {
			list-style: none;
			margin: 0;
			padding: 0;
			width: 100%;
			height: 100%;
		}

		:global(.container) {
			width: 100% !important;
			height: 100% !important;
			min-width: unset !important;
		}

		// Disable the zoom/scale effect on hover for chosen poster
		:global(.active .container) {
			transform: none !important;
			z-index: 1 !important;
		}
	}

	.chosen-info {
		flex: 1;
		position: relative;
		z-index: 1;

		h4 {
			margin: 0 0 var(--space-sm) 0;
			font-size: 2rem;
			color: var(--text);
			font-weight: 800;
			line-height: 1.1;
		}

		.chosen-year {
			font-size: 1rem;
			color: var(--text-muted);
			font-weight: 600;
			margin-bottom: var(--space-md);
			background: var(--background-secondary);
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-md);
			display: inline-block;
		}

		.chosen-overview {
			color: var(--text-muted);
			line-height: 1.5;
			margin: 0;
			padding: var(--space-md);
			background: var(--background-secondary);
			border-radius: var(--radius-md);
			border-left: 3px solid var(--accent, #8b5cf6);
			font-style: italic;
		}
	}

	.voting-summary {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-md);
		border-left: 4px solid var(--primary);

		h4 {
			margin: 0 0 var(--space-md) 0;
			font-weight: 600;
			color: var(--text);
			font-size: 1.125rem;
		}

		.summary-stats {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
			gap: var(--space-md);
		}

		.summary-stat {
			text-align: center;
			padding: var(--space-md);
			background: var(--background-secondary);
			border-radius: var(--radius-lg);
			border: 1px solid var(--border);
			box-shadow: var(--shadow-sm);
			transition: all 0.2s ease;

			&:hover {
				transform: translateY(-2px);
				box-shadow: var(--shadow-md);
			}

			.stat-value {
				display: block;
				font-size: 1.5rem;
				font-weight: 700;
				color: var(--primary);
				margin-bottom: var(--space-xs);
			}

			.stat-label {
				font-size: 0.875rem;
				color: var(--text-muted);
				font-weight: 500;
				text-transform: uppercase;
				letter-spacing: 0.025em;
			}
		}
	}

	@media (max-width: 768px) {
		.results-section {
			gap: var(--space-lg);
		}

		.section-header {
			margin-bottom: var(--space-md);

			h3 {
				font-size: 1.25rem;
			}

			p {
				font-size: 0.9rem;
			}
		}

		.winner-section {
			margin-bottom: var(--space-lg);
		}

		.winner-badge {
			padding: var(--space-xs) var(--space-md);
			font-size: 1rem;
			margin-bottom: var(--space-md);
		}

		.winner-card {
			flex-direction: column;
			text-align: center;
			padding: var(--space-lg);
			gap: var(--space-lg);

			.winner-poster {
				align-self: center;
				width: 160px;
				height: 240px;
			}

			.winner-details {
				h2 {
					font-size: 1.5rem;
				}

				.winner-stats {
					justify-content: center;
					gap: var(--space-md);

					.stat {
						min-width: 70px;

						.stat-value {
							font-size: 1.25rem;
						}

						.stat-label {
							font-size: 0.8rem;
						}
					}
				}

				.winner-overview {
					padding: var(--space-sm);
					font-size: 0.9rem;
				}
			}
		}

		.full-results {
			padding: var(--space-md);

			h4 {
				font-size: 1rem;
			}
		}

		.results-header,
		.result-row {
			grid-template-columns: 1fr;
			gap: var(--space-sm);
			text-align: center;
			padding: var(--space-sm);

			.rank,
			.votes,
			.score {
				font-size: 0.85rem;
			}
		}

		.result-row {
			.movie .movie-info {
				h5 {
					font-size: 0.85rem;
				}

				.release-year {
					font-size: 0.75rem;
				}
			}

			.breakdown .vote-breakdown {
				justify-content: center;

				.choice {
					font-size: 0.7rem;
					padding: 2px var(--space-xs);
				}
			}
		}

		.result-row .movie {
			justify-content: center;
		}

		.voting-summary {
			padding: var(--space-md);

			h4 {
				font-size: 1rem;
			}

			.summary-stats {
				grid-template-columns: 1fr;
				gap: var(--space-sm);
			}

			.summary-stat {
				padding: var(--space-sm);

				.stat-value {
					font-size: 1.25rem;
				}

				.stat-label {
					font-size: 0.8rem;
				}
			}
		}

		.no-votes {
			padding: var(--space-xl) var(--space-sm);

			:global(svg) {
				font-size: 2.5rem;
			}
		}

		.cycle-ratings-section {
			padding: var(--space-md);
		}

		.club-average-rating {
			padding: var(--space-sm);

			.average-label {
				font-size: 1rem;
			}

			.average-value {
				font-size: 1.25rem;

				.golden-asterisk {
					font-size: 32px;
					margin-right: 0.05em;
					transform: translateY(0.32em);
				}
			}
		}

		.ratings-header {
			padding: var(--space-sm);
			font-size: 0.9rem;
		}

		.ratings-content {
			padding: var(--space-sm);
			gap: var(--space-sm);
		}

		.rating-item {
			padding: var(--space-sm);
		}

		.rating-user {
			.username {
				font-size: 0.85rem;
			}

			.rating-score {
				font-size: 0.8rem;
				padding: 2px var(--space-xs);
			}

			.thoughts-only {
				font-size: 0.8rem;
				padding: 2px var(--space-xs);
			}
		}

		.thoughts-toggle {
			padding: 6px var(--space-xs);
			font-size: 0.8rem;
		}

		.thoughts-content {
			padding: var(--space-xs);

			p {
				font-size: 0.85rem;
			}
		}
	}
</style>
