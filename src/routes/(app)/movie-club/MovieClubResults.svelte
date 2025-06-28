<script lang="ts">
	import type { MovieClubCycleResponse } from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";

	export let cycleData: MovieClubCycleResponse;

	$: results = cycleData.voteResults || [];
	$: winner = results.length > 0 ? results[0] : null;
	$: hasVotes = results.some(r => r.totalVotes > 0);
</script>

<div class="results-section">
	<div class="section-header">
		<h3>
			<Icon icon="play" />
			Voting Results
		</h3>
		<p>Here are the results from this cycle's voting!</p>
	</div>

	{#if !hasVotes}
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
						<Poster media={winner.content} showRating={false} disableInteraction={true} />
						<div class="winner-overlay">
							<Icon icon="play" />
						</div>
					</div>
					
					<div class="winner-details">
						<h2>{winner.content.title}</h2>
						<div class="winner-stats">
							<div class="stat">
								<span class="stat-value">{winner.totalVotes}</span>
								<span class="stat-label">Total Votes</span>
							</div>
							<div class="stat">
								<span class="stat-value">{winner.weightedScore.toFixed(1)}</span>
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
							{#if index === 0}
								<Icon icon="sparkles" />
							{:else}
								#{index + 1}
							{/if}
						</div>
						
						<div class="movie">
							<div class="movie-poster">
								<Poster media={result.content} showRating={false} disableInteraction={true} />
							</div>
							<div class="movie-info">
								<h5>{result.content.title}</h5>
								<span class="release-year">
									{result.content.release_date ? new Date(result.content.release_date).getFullYear() : "Unknown"}
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
					<span class="stat-value">{results.reduce((sum, r) => sum + r.totalVotes, 0)}</span>
					<span class="stat-label">Total Votes Cast</span>
				</div>
				<div class="summary-stat">
					<span class="stat-value">{results.length}</span>
					<span class="stat-label">Movies Nominated</span>
				</div>
				<div class="summary-stat">
					<span class="stat-value">{results.filter(r => r.totalVotes > 0).length}</span>
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
		gap: 2rem;
	}

	.section-header {
		text-align: center;

		h3 {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 0.5rem;
			margin-bottom: 0.5rem;
			font-size: 1.5rem;
		}

		p {
			color: var(--text-muted);
			margin: 0;
		}
	}

	.no-votes {
		text-align: center;
		padding: 3rem 1rem;
		color: var(--text-muted);

		:global(svg) {
			font-size: 3rem;
			margin-bottom: 1rem;
			opacity: 0.5;
		}
	}

	.winner-section {
		position: relative;
		margin-bottom: 2rem;
	}

	.winner-badge {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		background: linear-gradient(135deg, #ffd700, #ffed4e);
		color: #000;
		padding: 0.75rem 1.5rem;
		border-radius: 20px;
		font-weight: 700;
		font-size: 1.1rem;
		width: fit-content;
		margin: 0 auto 1.5rem auto;
		box-shadow: 0 4px 12px rgba(255, 215, 0, 0.3);

		:global(svg) {
			font-size: 1.25rem;
		}
	}

	.winner-card {
		background: linear-gradient(135deg, rgba(255, 215, 0, 0.1), rgba(255, 237, 78, 0.05));
		border: 2px solid #ffd700;
		border-radius: 16px;
		padding: 2rem;
		display: flex;
		gap: 2rem;
		align-items: center;
		box-shadow: 0 8px 24px rgba(255, 215, 0, 0.2);
	}

	.winner-poster {
		position: relative;
		width: 150px;
		height: 225px;
		flex-shrink: 0;

		.winner-overlay {
			position: absolute;
			inset: 0;
			background: rgba(255, 215, 0, 0.9);
			display: flex;
			align-items: center;
			justify-content: center;
			color: #000;
			font-size: 3rem;
			border-radius: 8px;
		}
	}

	.winner-details {
		flex: 1;

		h2 {
			margin: 0 0 1rem 0;
			font-size: 2rem;
			color: var(--text);
		}

		.winner-stats {
			display: flex;
			gap: 2rem;
			margin-bottom: 1.5rem;

			.stat {
				text-align: center;

				.stat-value {
					display: block;
					font-size: 1.5rem;
					font-weight: 700;
					color: #ffd700;
				}

				.stat-label {
					font-size: 0.875rem;
					color: var(--text-muted);
				}
			}
		}

		.winner-overview {
			color: var(--text-muted);
			line-height: 1.5;
			margin: 0;
		}
	}

	.full-results {
		h4 {
			margin-bottom: 1rem;
			padding-bottom: 0.5rem;
			border-bottom: 1px solid var(--border);
		}
	}

	.results-table {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.results-header {
		display: grid;
		grid-template-columns: 60px 1fr 80px 80px 200px;
		gap: 1rem;
		padding: 1rem;
		background: var(--background-secondary);
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.result-row {
		display: grid;
		grid-template-columns: 60px 1fr 80px 80px 200px;
		gap: 1rem;
		padding: 1rem;
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 8px;
		align-items: center;

		&.winner {
			background: linear-gradient(135deg, rgba(255, 215, 0, 0.1), rgba(255, 237, 78, 0.05));
			border-color: #ffd700;
		}

		.rank {
			text-align: center;
			font-weight: 700;
			color: var(--primary);

			:global(svg) {
				color: #ffd700;
			}
		}

		.movie {
			display: flex;
			align-items: center;
			gap: 1rem;

			.movie-poster {
				width: 40px;
				height: 60px;
				flex-shrink: 0;
			}

			.movie-info {
				h5 {
					margin: 0 0 0.25rem 0;
					font-size: 0.9rem;
					line-height: 1.2;
				}

				.release-year {
					font-size: 0.8rem;
					color: var(--text-muted);
				}
			}
		}

		.votes, .score {
			text-align: center;
			font-weight: 600;
		}

		.breakdown {
			.vote-breakdown {
				display: flex;
				flex-wrap: wrap;
				gap: 0.5rem;

				.choice {
					padding: 0.25rem 0.5rem;
					border-radius: 4px;
					font-size: 0.75rem;
					font-weight: 500;

					&.first {
						background: rgba(34, 197, 94, 0.2);
						color: var(--success);
					}

					&.second {
						background: rgba(245, 158, 11, 0.2);
						color: var(--warning);
					}

					&.third {
						background: rgba(239, 68, 68, 0.2);
						color: var(--danger);
					}
				}

				.no-votes {
					color: var(--text-muted);
					font-size: 0.75rem;
					font-style: italic;
				}
			}
		}
	}

	.voting-summary {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 1.5rem;

		h4 {
			margin: 0 0 1rem 0;
		}

		.summary-stats {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
			gap: 1rem;
		}

		.summary-stat {
			text-align: center;
			padding: 1rem;
			background: var(--background-secondary);
			border-radius: 8px;

			.stat-value {
				display: block;
				font-size: 1.5rem;
				font-weight: 700;
				color: var(--primary);
			}

			.stat-label {
				font-size: 0.875rem;
				color: var(--text-muted);
			}
		}
	}

	@media (max-width: 768px) {
		.winner-card {
			flex-direction: column;
			text-align: center;
			padding: 1.5rem;

			.winner-poster {
				align-self: center;
			}

			.winner-details .winner-stats {
				justify-content: center;
			}
		}

		.results-header, .result-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
			text-align: center;
		}

		.result-row .movie {
			justify-content: center;
		}

		.voting-summary .summary-stats {
			grid-template-columns: 1fr;
		}
	}
</style>