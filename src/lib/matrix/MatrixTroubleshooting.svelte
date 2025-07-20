<script lang="ts">
	let { onClose }: { onClose: () => void } = $props();

	interface TroubleshootingSection {
		title: string;
		problems: TroubleshootingProblem[];
	}

	interface TroubleshootingProblem {
		issue: string;
		symptoms: string[];
		solutions: string[];
		links?: Array<{ text: string; url: string }>;
	}

	const troubleshootingSections: TroubleshootingSection[] = [
		{
			title: "Connection Issues",
			problems: [
				{
					issue: "Cannot connect to Matrix server",
					symptoms: [
						"Connection test fails",
						"'Failed to create Matrix client' error",
						"Network timeout errors",
					],
					solutions: [
						"Verify the server URL is correct and accessible",
						"Check that Matrix server (Synapse/Dendrite) is running",
						"Ensure firewall allows connections on the Matrix port",
						"Test server accessibility from Watcharr server: curl -I <server-url>",
						"Check server logs for error details",
					],
				},
				{
					issue: "Authentication fails with valid token",
					symptoms: [
						"'Failed to authenticate' error",
						"Access token appears correct but validation fails",
					],
					solutions: [
						"Verify the admin token has not expired",
						"Check that the token has admin privileges",
						"Regenerate admin access token from Matrix server",
						"Ensure token format is correct (no extra spaces or newlines)",
						"For Synapse: Use admin API token; For Dendrite: Use admin access token",
					],
				},
			],
		},
		{
			title: "Permission Issues",
			problems: [
				{
					issue: "Admin permissions insufficient",
					symptoms: [
						"Room creation fails",
						"User creation fails",
						"'Admin permissions' check fails",
					],
					solutions: [
						"Verify the admin user exists in Matrix server",
						"Check admin user has server admin privileges",
						"Review Matrix server configuration for admin settings",
						"For Dendrite: Use dendrite-admin-tool create-account <username>",
						"For Synapse: Use registration_shared_secret or admin API",
					],
					links: [
						{
							text: "Matrix Admin Documentation",
							url: "https://matrix-org.github.io/synapse/latest/admin_api/",
						},
					],
				},
			],
		},
		{
			title: "Room Creation Issues",
			problems: [
				{
					issue: "Cannot create test rooms",
					symptoms: [
						"Room creation validation fails",
						"'Failed to create test room' error",
						"Rooms created but not accessible",
					],
					solutions: [
						"Check Matrix server disk space and resources",
						"Verify room alias format is valid",
						"Review Matrix server logs for room creation errors",
						"Check server configuration allows room creation",
						"Ensure federation is disabled for local-only rooms",
					],
				},
				{
					issue: "Room aliases not working",
					symptoms: [
						"Room created but alias not accessible",
						"Alias conflicts or format errors",
					],
					solutions: [
						"Verify server name configuration matches Matrix server setup",
						"Check for existing alias conflicts",
						"Review alias format: #roomname:servername",
						"Ensure server name matches Matrix server's configured domain",
					],
				},
			],
		},
		{
			title: "Configuration Issues",
			problems: [
				{
					issue: "Server name mismatch",
					symptoms: [
						"User IDs not in expected format",
						"Room aliases not working",
						"Federation issues (if enabled)",
					],
					solutions: [
						"Match server name to Matrix server's configured server_name",
						"Update Watcharr Matrix server name setting",
						"Restart Watcharr after configuration changes",
						"Check Matrix server key configuration",
					],
				},
				{
					issue: "Invalid admin user ID format",
					symptoms: [
						"Admin user operations fail",
						"User ID format errors in logs",
					],
					solutions: [
						"Ensure admin user ID format: @username:servername",
						"Match servername to Matrix server configuration",
						"Create admin user if it doesn't exist",
						"Verify case sensitivity in usernames",
					],
				},
			],
		},
		{
			title: "Matrix Server-Specific Issues",
			problems: [
				{
					issue: "Synapse startup failures",
					symptoms: [
						"Matrix server not responding",
						"Connection refused errors",
						"Service not running",
					],
					solutions: [
						"Check Synapse service status: systemctl status matrix-synapse",
						"Review homeserver.yaml configuration file for syntax errors",
						"Check Synapse logs: journalctl -u matrix-synapse -f",
						"Verify database connectivity (PostgreSQL/SQLite)",
						"Ensure required ports are available (8008, 8448)",
					],
					links: [
						{
							text: "Synapse Installation Guide",
							url: "https://matrix-org.github.io/synapse/latest/setup/installation.html",
						},
					],
				},
				{
					issue: "Dendrite startup failures",
					symptoms: [
						"Matrix server not responding",
						"Connection refused errors",
						"Service not running",
					],
					solutions: [
						"Check Dendrite service status: systemctl status dendrite",
						"Review dendrite.yaml configuration file for syntax errors",
						"Check Dendrite logs: journalctl -u dendrite -f",
						"Verify database connectivity (PostgreSQL/SQLite)",
						"Ensure required ports are available (8008, 8448)",
					],
					links: [
						{
							text: "Dendrite Installation Guide",
							url: "https://matrix-org.github.io/dendrite/installation",
						},
					],
				},
				{
					issue: "Database connectivity issues",
					symptoms: [
						"Matrix server fails to start",
						"Database connection errors in logs",
					],
					solutions: [
						"Verify database server is running and accessible",
						"Check database credentials in Matrix server config",
						"Test database connection independently",
						"Review database permissions for Matrix server user",
						"Check database schema is properly initialized",
						"For Synapse: Check PostgreSQL config in homeserver.yaml",
						"For Dendrite: Check database config in dendrite.yaml",
					],
				},
			],
		},
	];

	let expandedSections = $state<Record<string, boolean>>({});
	let expandedProblems = $state<Record<string, boolean>>({});

	function toggleSection(title: string) {
		expandedSections[title] = !expandedSections[title];
	}

	function toggleProblem(issue: string) {
		expandedProblems[issue] = !expandedProblems[issue];
	}
</script>

<div class="modal-backdrop" onclick={onClose}>
	<div class="modal" onclick={(e) => e.stopPropagation()}>
		<div class="modal-header">
			<h2>Matrix Setup Troubleshooting Guide</h2>
			<button class="close-button" onclick={onClose}>×</button>
		</div>

		<div class="modal-content">
			<div class="guide-intro">
				<p>
					This guide helps resolve common Matrix server integration issues.
					Click on sections and problems to expand detailed solutions.
				</p>
			</div>

			<div class="troubleshooting-sections">
				{#each troubleshootingSections as section}
					<div class="section">
						<div
							class="section-header"
							onclick={() => toggleSection(section.title)}
						>
							<h3>{section.title}</h3>
							<span class="toggle-icon"
								>{expandedSections[section.title] ? "▼" : "▶"}</span
							>
						</div>

						{#if expandedSections[section.title]}
							<div class="section-content">
								{#each section.problems as problem}
									<div class="problem">
										<div
											class="problem-header"
											onclick={() => toggleProblem(problem.issue)}
										>
											<h4>{problem.issue}</h4>
											<span class="toggle-icon"
												>{expandedProblems[problem.issue] ? "▼" : "▶"}</span
											>
										</div>

										{#if expandedProblems[problem.issue]}
											<div class="problem-content">
												<div class="symptoms">
													<h5>Symptoms:</h5>
													<ul>
														{#each problem.symptoms as symptom}
															<li>{symptom}</li>
														{/each}
													</ul>
												</div>

												<div class="solutions">
													<h5>Solutions:</h5>
													<ol>
														{#each problem.solutions as solution}
															<li>{solution}</li>
														{/each}
													</ol>
												</div>

												{#if problem.links}
													<div class="helpful-links">
														<h5>Helpful Links:</h5>
														<ul class="links-list">
															{#each problem.links as link}
																<li>
																	<a
																		href={link.url}
																		target="_blank"
																		rel="noopener noreferrer"
																	>
																		{link.text} ↗
																	</a>
																</li>
															{/each}
														</ul>
													</div>
												{/if}
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/each}
			</div>

			<div class="additional-help">
				<h3>Still Need Help?</h3>
				<p>If you're still experiencing issues:</p>
				<ul>
					<li>Check Watcharr server logs for detailed error messages</li>
					<li>Review Matrix server logs for issues</li>
					<li>
						Consult the official Matrix server documentation:
						<a
							href="https://matrix-org.github.io/synapse/latest/"
							target="_blank"
							rel="noopener noreferrer">Synapse ↗</a
						>
						|
						<a
							href="https://matrix-org.github.io/dendrite/"
							target="_blank"
							rel="noopener noreferrer">Dendrite ↗</a
						>
					</li>
					<li>
						Report issues to the <a
							href="https://github.com/sbondCo/Watcharr/issues"
							target="_blank"
							rel="noopener noreferrer">Watcharr GitHub repository ↗</a
						>
					</li>
				</ul>
			</div>
		</div>

		<div class="modal-footer">
			<button onclick={onClose} class="close-action">Close</button>
		</div>
	</div>
</div>

<style lang="scss">
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal {
		background: var(--bg-color);
		border-radius: 12px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
		width: 90%;
		max-width: 900px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 20px;
		border-bottom: 1px solid var(--accent-color);

		h2 {
			margin: 0;
			font-size: 1.5em;
		}

		.close-button {
			background: none;
			border: none;
			font-size: 24px;
			cursor: pointer;
			color: var(--text-color);
			padding: 5px;
			border-radius: 50%;
			width: 35px;
			height: 35px;
			display: flex;
			align-items: center;
			justify-content: center;

			&:hover {
				background: var(--accent-color);
			}
		}
	}

	.modal-content {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}

	.guide-intro {
		margin-bottom: 25px;
		padding: 15px;
		background: var(--bg-color);
		border-radius: 8px;
		border-left: 4px solid var(--accent-color);

		p {
			margin: 0;
			line-height: 1.6;
		}
	}

	.troubleshooting-sections {
		display: flex;
		flex-direction: column;
		gap: 15px;
	}

	.section {
		border: 1px solid var(--accent-color);
		border-radius: 8px;
		overflow: hidden;
	}

	.section-header {
		background: var(--bg-color);
		padding: 15px;
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: background-color 0.2s ease;

		&:hover {
			background: var(--accent-color);
		}

		h3 {
			margin: 0;
			font-size: 1.2em;
			color: var(--accent-color);
		}

		.toggle-icon {
			font-size: 1.1em;
			color: var(--text-color-accent);
		}
	}

	.section-content {
		padding: 0;
	}

	.problem {
		border-top: 1px solid var(--accent-color);

		&:first-child {
			border-top: none;
		}
	}

	.problem-header {
		padding: 12px 15px;
		cursor: pointer;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: background-color 0.2s ease;

		&:hover {
			background: var(--bg-color);
		}

		h4 {
			margin: 0;
			font-size: 1.05em;
			font-weight: 500;
		}

		.toggle-icon {
			font-size: 0.9em;
			color: var(--text-color-accent);
		}
	}

	.problem-content {
		padding: 15px;
		background: var(--bg-color);
		border-top: 1px solid var(--accent-color);

		h5 {
			margin: 0 0 8px 0;
			font-size: 0.95em;
			font-weight: 600;
			color: var(--accent-color);
		}

		ul,
		ol {
			margin: 0 0 15px 0;
			padding-left: 20px;

			li {
				margin-bottom: 4px;
				line-height: 1.4;
			}
		}

		.symptoms,
		.solutions,
		.helpful-links {
			margin-bottom: 15px;

			&:last-child {
				margin-bottom: 0;
			}
		}

		.links-list {
			list-style: none;
			padding-left: 0;

			li {
				margin-bottom: 8px;
			}

			a {
				color: var(--accent-color);
				text-decoration: none;
				border-bottom: 1px solid transparent;
				transition: border-bottom-color 0.2s ease;

				&:hover {
					border-bottom-color: var(--accent-color);
				}
			}
		}
	}

	.additional-help {
		margin-top: 30px;
		padding: 20px;
		background: var(--bg-color);
		border-radius: 8px;
		border-left: 4px solid var(--warning-color, #f59e0b);

		h3 {
			margin: 0 0 10px 0;
			font-size: 1.2em;
		}

		p {
			margin: 0 0 10px 0;
			line-height: 1.5;
		}

		ul {
			margin: 0;
			padding-left: 20px;

			li {
				margin-bottom: 5px;
				line-height: 1.4;
			}
		}

		a {
			color: var(--accent-color);
			text-decoration: none;
			border-bottom: 1px solid transparent;
			transition: border-bottom-color 0.2s ease;

			&:hover {
				border-bottom-color: var(--accent-color);
			}
		}
	}

	.modal-footer {
		padding: 20px;
		border-top: 1px solid var(--accent-color);
		background: var(--bg-color);
		display: flex;
		justify-content: flex-end;

		button.primary {
			background: var(--accent-color);
			color: white;
			border: none;
			padding: 10px 20px;
			border-radius: 6px;
			font-weight: 500;
			cursor: pointer;
			transition: background-color 0.2s ease;

			&:hover {
				background: var(--accent-color-dark);
			}
		}

		button.close-action {
			background: var(--bg-color);
			border: 1px solid var(--border-color);
			color: var(--text-muted, #666);
			padding: 10px 20px;
			border-radius: 6px;
			font-size: 0.9rem;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s ease;

			&:hover {
				color: var(--text-color);
				background: var(--bg-secondary, rgba(0, 0, 0, 0.02));
				border-color: var(--text-muted, #666);
				transform: translateY(-1px);
			}

			&:focus {
				outline: none;
				box-shadow: 0 0 0 2px var(--accent-color);
			}
		}
	}

	@media (max-width: 768px) {
		.modal {
			width: 95%;
			max-height: 95vh;
		}

		.modal-content {
			padding: 15px;
		}

		.modal-header {
			padding: 15px;
		}

		.section-header,
		.problem-header {
			padding: 12px;
		}

		.problem-content {
			padding: 12px;
		}
	}
</style>
