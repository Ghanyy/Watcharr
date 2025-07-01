<script lang="ts">
	import { onMount } from "svelte";
	import { getArchivedMovieClubCycles } from "@/lib/util/api";
	import type { MovieClubCycleResponse } from "@/types";
	import Spinner from "@/lib/Spinner.svelte";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";

	let archivedCycles: MovieClubCycleResponse[] = [];
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		try {
			archivedCycles = await getArchivedMovieClubCycles();
		} catch (err: any) {
			console.error("Failed to fetch archived cycles:", err);
			error = err.response?.data?.error || "Failed to load archived cycles";
		} finally {
			loading = false;
		}
	});

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString(undefined, {
			year: "numeric",
			month: "long",
			day: "numeric",
		});
	}

	function goBack() {
		window.location.href = "/movie-club";
	}
</script>

<svelte:head>
	<title>Movie Club Archives - Watcharr</title>
</svelte:head>

<div class="archives-page">
	<header>
		<h1>🗂️ Movie Club Archives</h1>
		<p>Browse through all completed movie club cycles</p>
	</header>

	<button class="back-btn" on:click={goBack}>
		<Icon icon="arrow" />
		Back to Movie Club
	</button>

	{#if loading}
		<div class="loading">
			<Spinner />
			<p>Loading archived cycles...</p>
		</div>
	{:else if error}
		<Error {error} />
	{:else if archivedCycles.length === 0}
		<div class="empty-state">
			<Icon icon="archive" />
			<h3>No Archives Yet</h3>
			<p>
				There are no completed movie club cycles to show. Archives will appear
				here once cycles have finished their full lifecycle.
			</p>
		</div>
	{:else}
		<div class="archives-list">
			{#each archivedCycles as cycle}
				<div class="archive-card">
					<div class="cycle-header">
						<h3>{cycle.cycle.name}</h3>
						<div class="cycle-meta">
							<span class="cycle-dates">
								{formatDate(cycle.cycle.phaseStartDate)} - {formatDate(
									cycle.cycle.watchingEndDate,
								)}
							</span>
							<span class="phase-badge completed">Completed</span>
						</div>
					</div>

					{#if cycle.cycle.description}
						<p class="cycle-description">{cycle.cycle.description}</p>
					{/if}

					{#if cycle.voteResults && cycle.voteResults.length > 0}
						<div class="results-section">
							<h4>🏆 Final Results</h4>
							<div class="results-list">
								{#each cycle.voteResults as result, index}
									<div class="result-item" class:winner={index === 0}>
										<div class="place">
											{#if index === 0}
												🏆
											{:else if index === 1}
												🥈
											{:else if index === 2}
												🥉
											{:else}
												{index + 1}.
											{/if}
										</div>
										<div class="movie-info">
											<strong>{result.content.title}</strong>
											{#if result.content.release_date}
												<span class="release-year">
													({new Date(result.content.release_date).getFullYear()})
												</span>
											{/if}
										</div>
										<div class="score-info">
											<span class="weighted-score">{result.weightedScore} pts</span>
											<span class="vote-details">{result.totalVotes} votes</span>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{:else if cycle.cycle.winnerContent}
						<div class="winner-section">
							<h4>🏆 Winner</h4>
							<div class="winner-movie">
								<strong>{cycle.cycle.winnerContent.title}</strong>
								{#if cycle.cycle.winnerContent.release_date}
									<span class="release-year">
										({new Date(
											cycle.cycle.winnerContent.release_date,
										).getFullYear()})
									</span>
								{/if}
							</div>
						</div>
					{/if}

					<div class="cycle-stats">
						<div class="stat">
							<span class="stat-value">{cycle.cycle.nominations?.length || 0}</span>
							<span class="stat-label">Nominations</span>
						</div>
						<div class="stat">
							<span class="stat-value">{cycle.totalVoters || 0}</span>
							<span class="stat-label">Voters</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style lang="scss">
	.archives-page {
		min-height: 100vh;
		padding: var(--space-xl);
		background: var(--background);

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

	.back-btn {
		display: flex;
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
		transition: all 0.2s ease;
		box-shadow: var(--shadow-sm);
		margin: 0 auto var(--space-xl) auto;
		max-width: 800px;
		width: fit-content;

		&:hover {
			background: var(--background-secondary);
			transform: translateY(-1px);
			box-shadow: var(--shadow-md);
		}

		&:focus {
			outline: none;
			box-shadow:
				var(--shadow-md),
				0 0 0 3px rgba(59, 130, 246, 0.2);
		}

		:global(svg) {
			transform: rotate(180deg);
		}
	}

	header {
		text-align: center;
		margin-bottom: var(--space-2xl);
		max-width: 800px;
		margin-left: auto;
		margin-right: auto;

		h1 {
			margin: 0 0 var(--space-sm) 0;
			font-size: 2.5rem;
			font-weight: 700;
			color: var(--text);
			line-height: 1.2;
		}

		p {
			color: var(--text-muted);
			margin: 0;
			font-size: 1.1rem;
			line-height: 1.4;
		}
	}

	.loading {
		text-align: center;
		padding: var(--space-2xl);
		color: var(--text-muted);

		p {
			margin-top: var(--space-md);
			font-size: 0.95rem;
		}
	}

	.empty-state {
		text-align: center;
		padding: var(--space-2xl) var(--space-md);
		color: var(--text-muted);
		border-radius: var(--radius-lg);
		border: 2px dashed var(--border);
		margin: var(--space-xl) auto;
		max-width: 500px;

		:global(svg) {
			font-size: 4rem;
			margin-bottom: var(--space-lg);
			opacity: 0.4;
		}

		h3 {
			margin: 0 0 var(--space-md) 0;
			font-size: 1.25rem;
			color: var(--text);
		}

		p {
			margin: 0;
			line-height: 1.5;
			font-size: 0.95rem;
		}
	}

	.archives-list {
		display: grid;
		gap: var(--space-lg);
		max-width: 800px;
		margin: 0 auto;
	}

	.archive-card {
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		box-shadow: var(--shadow-sm);
		transition: all 0.2s ease;

		&:hover {
			transform: translateY(-2px);
			box-shadow: var(--shadow-md);
		}
	}

	.cycle-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-md);
		gap: var(--space-md);

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
			color: var(--text);
			line-height: 1.2;
		}

		.cycle-meta {
			display: flex;
			flex-direction: column;
			align-items: flex-end;
			gap: var(--space-xs);
		}

		.cycle-dates {
			font-size: 0.85rem;
			color: var(--text-muted);
			white-space: nowrap;
		}

		.phase-badge {
			padding: var(--space-xs) var(--space-sm);
			border-radius: var(--radius-sm);
			font-size: 0.75rem;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.025em;

			&.completed {
				background: var(--success, #22c55e);
				color: white;
			}
		}
	}

	.cycle-description {
		color: var(--text-muted);
		margin: 0 0 var(--space-md) 0;
		line-height: 1.5;
		font-size: 0.95rem;
	}

	.results-section {
		margin: var(--space-md) 0;
		padding: var(--space-md);
		background: var(--background-secondary);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--warning, #f59e0b);

		h4 {
			margin: 0 0 var(--space-md) 0;
			font-size: 0.9rem;
			font-weight: 600;
			color: var(--text);
			text-transform: uppercase;
			letter-spacing: 0.025em;
		}

		.results-list {
			display: flex;
			flex-direction: column;
			gap: var(--space-sm);
		}

		.result-item {
			display: flex;
			align-items: center;
			gap: var(--space-md);
			padding: var(--space-sm);
			background: var(--background);
			border-radius: var(--radius-md);
			border: 1px solid var(--border);
			transition: all 0.2s ease;

			&.winner {
				border-color: var(--warning, #f59e0b);
				background: var(--background);
				box-shadow: 0 0 0 1px rgba(245, 158, 11, 0.1);
			}

			.place {
				font-size: 1.1rem;
				font-weight: 700;
				min-width: 2rem;
				text-align: center;
				color: var(--text);
			}

			.movie-info {
				flex: 1;
				min-width: 0;

				strong {
					color: var(--text);
					font-weight: 600;
					display: block;
					line-height: 1.2;
				}

				.release-year {
					color: var(--text-muted);
					font-size: 0.85rem;
					margin-top: var(--space-xs);
					display: block;
				}
			}

			.score-info {
				text-align: right;
				display: flex;
				flex-direction: column;
				gap: 2px;

				.weighted-score {
					font-weight: 600;
					color: var(--primary);
					font-size: 0.9rem;
				}

				.vote-details {
					font-size: 0.8rem;
					color: var(--text-muted);
				}
			}
		}
	}

	.winner-section {
		margin: var(--space-md) 0;
		padding: var(--space-md);
		background: var(--background-secondary);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--warning, #f59e0b);

		h4 {
			margin: 0 0 var(--space-sm) 0;
			font-size: 0.9rem;
			font-weight: 600;
			color: var(--text);
			text-transform: uppercase;
			letter-spacing: 0.025em;
		}

		.winner-movie {
			strong {
				color: var(--text);
				font-weight: 600;
			}

			.release-year {
				color: var(--text-muted);
				font-size: 0.9rem;
				margin-left: var(--space-xs);
			}
		}
	}

	.cycle-stats {
		display: flex;
		gap: var(--space-lg);
		margin-top: var(--space-md);
		padding-top: var(--space-md);
		border-top: 1px solid var(--border);

		.stat {
			text-align: center;

			.stat-value {
				display: block;
				font-size: 1.25rem;
				font-weight: 700;
				color: var(--primary);
				line-height: 1;
			}

			.stat-label {
				display: block;
				font-size: 0.8rem;
				color: var(--text-muted);
				text-transform: uppercase;
				letter-spacing: 0.025em;
				margin-top: var(--space-xs);
			}
		}
	}

	@media (max-width: 768px) {
		.archives-page {
			padding: var(--space-lg) var(--space-md);
		}

		.back-btn {
			margin-bottom: var(--space-md);
		}

		header {
			h1 {
				font-size: 2rem;
			}

			p {
				font-size: 1rem;
			}
		}

		.cycle-header {
			flex-direction: column;
			align-items: flex-start;

			.cycle-meta {
				align-items: flex-start;
				width: 100%;
			}
		}

		.cycle-stats {
			justify-content: space-around;
		}

		.results-section {
			.result-item {
				flex-direction: column;
				align-items: flex-start;
				gap: var(--space-sm);
				text-align: left;

				.place {
					align-self: flex-start;
					min-width: auto;
				}

				.movie-info {
					order: 1;
					width: 100%;

					strong {
						font-size: 0.95rem;
					}

					.release-year {
						font-size: 0.8rem;
						margin-top: 2px;
					}
				}

				.score-info {
					order: 2;
					align-self: flex-start;
					text-align: left;
					flex-direction: row;
					gap: var(--space-sm);
					width: 100%;
					justify-content: space-between;

					.weighted-score {
						font-size: 0.85rem;
					}

					.vote-details {
						font-size: 0.75rem;
					}
				}
			}
		}
	}
</style>