<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import PageError from "@/lib/PageError.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import type {
		Content,
		RadarrSettings,
		ServerConfig,
		SonarrSettings,
		DropDownItem,
		MovieClubSettings,
		MatrixSettings,
		MatrixServerType,
		AppServiceSettings,
	} from "@/types";
	import axios from "axios";
	import SonarrModal from "./modals/SonarrModal.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingButton from "@/lib/settings/SettingButton.svelte";
	import RadarrModal from "./modals/RadarrModal.svelte";
	import { getServerFeatures } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import Stats from "@/lib/stats/Stats.svelte";
	import Error from "@/lib/Error.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import TwitchModal from "./modals/TwitchModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import TaskScheduleModal from "./modals/TaskScheduleModal.svelte";
	import TrustedHeaderAuthModal from "./modals/TrustedHeaderAuthModal.svelte";
	import MatrixValidation from "@/lib/matrix/MatrixValidation.svelte";

	let serverConfig: ServerConfig | undefined = $state();
	let jellyfinOrEmby = $derived(serverConfig?.USE_EMBY ? "Emby" : "Jellyfin");
	let sonarrModalOpen = $state(false);
	let sonarrServerEditing: SonarrSettings | undefined = $state();
	let sonarrModalEditing = $state(false);
	let radarrModalOpen = $state(false);
	let radarrServerEditing: RadarrSettings | undefined = $state();
	let radarrModalEditing = $state(false);
	let twitchModalOpen = $state(false);
	let taskScheduleModalOpen = $state(false);
	let headerSSOModalOpen = $state(false);
	// Disabled vars for disabling inputs until api request completes
	let signupDisabled = $state(false);
	let debugDisabled = $state(false);
	let jfDisabled = $state(false);
	let tmdbkDisabled = $state(false);
	let plexHostDisabled = $state(false);
	let countryDisabled = $state(false);
	let useEmbyDisabled = $state(false);
	// Movie Club disabled vars
	let movieClubEnabledDisabled = $state(false);
	let movieClubCommunityDisabled = $state(false);
	let movieClubNominationsDisabled = $state(false);
	let movieClubVotesDisabled = $state(false);
	let movieClubDurationDisabled = $state(false);

	// Matrix disabled vars
	let matrixEnabledDisabled = $state(false);
	let matrixServerUrlDisabled = $state(false);
	let matrixServerTypeDisabled = $state(false);
	let matrixServerNameDisabled = $state(false);
	let matrixAdminTokenDisabled = $state(false);
	let matrixAdminUserIdDisabled = $state(false);
	let matrixSpaceNameDisabled = $state(false);
	// Note: Registration secret removed - Matrix integration simplified to use Application Service
	let matrixTestLoading = $state(false);
	let matrixValidationModalOpen = $state(false);
	let createRoomsLoading = $state(false);

	async function getServerConfig() {
		serverConfig = (await axios.get(`/server/admin/config`))
			.data as ServerConfig;
	}

	export function updateServerConfig<K extends keyof ServerConfig>(
		name: K,
		value: ServerConfig[K],
		done?: (respData?: any) => void,
	) {
		if (!serverConfig) {
			console.error("updateServerConfig: No server config to update!");
			notify({ type: "error", text: "No server config to update!" });
			return;
		}
		console.log("Updating server setting", name, "to", value);
		const originalValue = serverConfig[name];
		const nid = notify({ type: "loading", text: "Updating" });
		let ep = "/server/admin/config";
		if (name === "PLEX_HOST") {
			ep = "/server/admin/config/plex_host";
		}
		axios
			.post(ep, { key: name, value: value })
			.then((r) => {
				if (r.status === 200) {
					serverConfig![name] = value;
					// Update store config as well
					if (store.config) {
						store.config = { ...store.config, [name]: value };
					}
					notify({ id: nid, type: "success", text: "Updated" });
					if (typeof done !== "undefined") done(r?.data);
				}
			})
			.catch((err) => {
				console.error("Failed to update user setting", err);
				notify({ id: nid, type: "error", text: "Couldn't Update" });
				serverConfig![name] = originalValue;
				if (typeof done !== "undefined") done();
			});
	}

	interface ServerStats {
		users: number;
		privateUsers: number;
		watchedMovies: number;
		watchedShows: number;
		watchedSeasons: number;
		mostWatchedMovie: Content;
		mostWatchedShow: Content;
		activities: number;
	}

	async function getServerStats() {
		return (await axios.get("/server/stats")).data as ServerStats;
	}

	function updateMovieClubConfig<K extends keyof MovieClubSettings>(
		field: K,
		value: MovieClubSettings[K],
		done?: () => void,
	) {
		if (!serverConfig) return;

		const updatedSettings = {
			...serverConfig.MOVIE_CLUB,
			[field]: value,
		};

		updateServerConfig("MOVIE_CLUB", updatedSettings, done);
	}

	function updateMatrixConfig<K extends keyof MatrixSettings>(
		field: K,
		value: MatrixSettings[K],
		done?: () => void,
	) {
		if (!serverConfig) return;

		const updatedMatrixSettings = {
			...serverConfig.MOVIE_CLUB.matrix,
			[field]: value,
		};

		const updatedMovieClubSettings = {
			...serverConfig.MOVIE_CLUB,
			matrix: updatedMatrixSettings,
		};

		updateServerConfig("MOVIE_CLUB", updatedMovieClubSettings, done);
	}
	
	function updateMatrixAppServiceConfig<K extends keyof AppServiceSettings>(
		field: K,
		value: AppServiceSettings[K],
		done?: () => void,
	) {
		if (!serverConfig) return;
		
		const updatedAppServiceSettings = {
			...serverConfig.MOVIE_CLUB.matrix.appService,
			[field]: value,
		};
		
		const updatedMatrixSettings = {
			...serverConfig.MOVIE_CLUB.matrix,
			appService: updatedAppServiceSettings,
		};
		
		const updatedMovieClubSettings = {
			...serverConfig.MOVIE_CLUB,
			matrix: updatedMatrixSettings,
		};

		updateServerConfig("MOVIE_CLUB", updatedMovieClubSettings, done);
	}

	async function testMatrixConnection() {
		if (!serverConfig?.MOVIE_CLUB?.matrix) return;

		matrixTestLoading = true;
		try {
			const response = await axios.post("/matrix/test-connection", {
				serverUrl: serverConfig.MOVIE_CLUB.matrix.serverUrl,
				adminToken: serverConfig.MOVIE_CLUB.matrix.adminToken,
			});

			if (response.status === 200) {
				notify({ type: "success", text: "Matrix connection successful!" });
			}
		} catch (error) {
			console.error("Matrix connection test failed:", error);
			notify({
				type: "error",
				text: "Matrix connection failed. Check server URL and admin token.",
			});
		} finally {
			matrixTestLoading = false;
		}
	}

	async function createRoomsForExistingCycles() {
		if (!serverConfig?.MOVIE_CLUB?.matrix?.enabled) {
			notify({ type: "error", text: "Matrix integration is not enabled" });
			return;
		}

		createRoomsLoading = true;
		try {
			const response = await axios.post(
				"/matrix/create-rooms-for-existing-cycles",
			);

			if (response.status === 200) {
				const result = response.data.result;
				if (result.totalCycles === 0) {
					notify({
						type: "info",
						text: "No active watching cycles found that need Matrix rooms",
					});
				} else if (result.createdRooms > 0) {
					let message = `Successfully created ${result.createdRooms} Matrix room(s)`;
					if (result.failedRooms > 0) {
						message += ` (${result.failedRooms} failed)`;
					}
					notify({ type: "success", text: message });
				} else {
					notify({
						type: "error",
						text: `Failed to create any rooms. ${result.errors?.[0] || "Unknown error"}`,
					});
				}
			} else if (response.status === 207) {
				// Partial success
				const result = response.data.result;
				notify({
					type: "warning",
					text: `Created ${result.createdRooms} rooms, but ${result.failedRooms} failed`,
				});
			}
		} catch (error: any) {
			console.error("Failed to create rooms for existing cycles:", error);
			let errorMessage = "Failed to create rooms for existing cycles";

			if (error.response?.status === 400) {
				errorMessage = "Matrix integration is not enabled";
			} else if (error.response?.status === 503) {
				errorMessage = "Matrix client is not initialized";
			} else if (error.response?.data?.error) {
				errorMessage = error.response.data.error;
			}

			notify({ type: "error", text: errorMessage });
		} finally {
			createRoomsLoading = false;
		}
	}

	async function downloadRegistrationFile() {
		if (!serverConfig?.MOVIE_CLUB?.matrix) return;

		try {
			const response = await axios.get("/matrix/registration-file", {
				responseType: "blob",
			});

			if (response.status === 200) {
				// Create download link
				const blob = new Blob([response.data], { type: "application/x-yaml" });
				const url = window.URL.createObjectURL(blob);
				const link = document.createElement("a");
				link.href = url;
				link.download = "watcharr-registration.yaml";
				document.body.appendChild(link);
				link.click();
				document.body.removeChild(link);
				window.URL.revokeObjectURL(url);

				notify({
					type: "success",
					text: "Registration file downloaded successfully",
				});
			}
		} catch (error: any) {
			console.error("Registration file download failed:", error);
			let errorMessage = "Failed to download registration file";

			if (error.response?.status === 400) {
				errorMessage = "Matrix integration is not enabled";
			} else if (error.response?.data?.error) {
				errorMessage = error.response.data.error;
			}

			notify({ type: "error", text: errorMessage });
		}
	}
</script>

<div class="content">
	<div class="inner">
		<SettingsList>
			<h2>Server Settings</h2>

			<Stats>
				{#await getServerStats()}
					<Spinner />
				{:then stats}
					<Stat name="Users" value={stats.users} href="/manage_users" large />
					<Stat name="Private Users" value={stats.privateUsers} large />
					<Stat name="Watched Movies" value={stats.watchedMovies} large />
					<Stat name="Watched Shows" value={stats.watchedShows} large />
					<Stat name="Watched Seasons" value={stats.watchedSeasons} large />
					<Stat name="Activities" value={stats.activities} large />
					{#if stats.mostWatchedMovie?.title}
						<Stat
							name="Most Watched Movie"
							value={stats.mostWatchedMovie.title}
							href="/movie/{stats.mostWatchedMovie.tmdbId}"
						/>
					{/if}
					{#if stats.mostWatchedShow?.title}
						<Stat
							name="Most Watched Show"
							value={stats.mostWatchedShow.title}
							href="/tv/{stats.mostWatchedShow.tmdbId}"
						/>
					{/if}
				{:catch err}
					<Error error={err} pretty="Failed to get server stats!" />
				{/await}
			</Stats>

			{#await getServerConfig()}
				<Spinner />
			{:then}
				<h3>General</h3>
				{#if serverConfig}
					<Setting
						title="Default Country"
						desc="Default country for new users. This can be changed per user and won't affect existing users."
					>
						<RegionDropDown
							selectedCountry={serverConfig.DEFAULT_COUNTRY}
							disabled={countryDisabled}
							onChange={(c) => {
								countryDisabled = true;
								updateServerConfig("DEFAULT_COUNTRY", c, () => {
									countryDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="{jellyfinOrEmby} Host"
						desc="Point to your {jellyfinOrEmby} server to enable related features. Don't change server after
        already using another."
					>
						<input
							type="text"
							placeholder="https://{jellyfinOrEmby.toLowerCase()}.example.com"
							bind:value={serverConfig.JELLYFIN_HOST}
							onblur={() => {
								jfDisabled = true;
								updateServerConfig(
									"JELLYFIN_HOST",
									serverConfig!.JELLYFIN_HOST,
									() => {
										jfDisabled = false;
									},
								);
							}}
							disabled={jfDisabled}
						/>
					</Setting>
					<Setting
						title="Use Emby"
						desc="Do you want to pretend you're using Emby instead of Jellyfin?"
						row
					>
						<Checkbox
							name="USE_EMBY"
							disabled={useEmbyDisabled}
							value={serverConfig.USE_EMBY}
							toggled={(on) => {
								useEmbyDisabled = true;
								updateServerConfig("USE_EMBY", on, () => {
									useEmbyDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="Plex Host"
						desc="Point to your Plex server to enable related features. Don't change server after
        already using another."
					>
						<input
							type="text"
							placeholder="https://plex.example.com"
							bind:value={serverConfig.PLEX_HOST}
							onblur={() => {
								plexHostDisabled = true;
								updateServerConfig(
									"PLEX_HOST",
									serverConfig!.PLEX_HOST,
									(rData) => {
										plexHostDisabled = false;
										serverConfig!.PLEX_MACHINE_ID = rData?.PLEX_MACHINE_ID;
									},
								);
							}}
							disabled={plexHostDisabled}
						/>
						{#if serverConfig.PLEX_MACHINE_ID}
							<span style="font-size: 10px"
								>Machine Id: {serverConfig.PLEX_MACHINE_ID}</span
							>
						{/if}
					</Setting>
					<Setting title="TMDB Key" desc="Provide your own TMDB API Key">
						<input
							type="password"
							placeholder="TMDB Key"
							bind:value={serverConfig.TMDB_KEY}
							onblur={() => {
								tmdbkDisabled = true;
								updateServerConfig("TMDB_KEY", serverConfig!.TMDB_KEY, () => {
									tmdbkDisabled = false;
								});
							}}
							disabled={tmdbkDisabled}
						/>
					</Setting>
					<Setting
						title="Signup"
						desc="Allow signing up with Watcharr credentials."
						row
					>
						<Checkbox
							name="SIGNUP_ENABLED"
							disabled={signupDisabled}
							value={serverConfig.SIGNUP_ENABLED}
							toggled={(on) => {
								signupDisabled = true;
								updateServerConfig("SIGNUP_ENABLED", on, () => {
									signupDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting title="Debug" desc="Enable debug logging." row>
						<Checkbox
							name="DEBUG"
							disabled={debugDisabled}
							value={serverConfig.DEBUG}
							toggled={(on) => {
								debugDisabled = true;
								updateServerConfig("DEBUG", on, () => {
									debugDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting>
						<SettingButton
							title="Task Schedule"
							desc="View and configure server task schedule."
							icon={"arrow"}
							onClick={() => {
								taskScheduleModalOpen = true;
							}}
						/>
					</Setting>
					{#if taskScheduleModalOpen}
						<TaskScheduleModal onClose={() => (taskScheduleModalOpen = false)}
						></TaskScheduleModal>
					{/if}
					<Setting>
						<SettingButton
							title="Trusted Header Authentication"
							desc="Configure trusted header single sign-on."
							icon={"arrow"}
							onClick={() => {
								headerSSOModalOpen = true;
							}}
						/>
					</Setting>
					{#if headerSSOModalOpen}
						<TrustedHeaderAuthModal onClose={() => (headerSSOModalOpen = false)}
						></TrustedHeaderAuthModal>
					{/if}

					<h3>Movie Club</h3>
					<Setting
						title="Enable Movie Club"
						desc="Allow users to participate in movie club cycles for group movie selection. After enabling, visit the Movie Club page to create your first cycle."
						row
					>
						<Checkbox
							name="MOVIE_CLUB_ENABLED"
							disabled={movieClubEnabledDisabled}
							value={serverConfig.MOVIE_CLUB.enabled}
							toggled={(on) => {
								movieClubEnabledDisabled = true;
								// If disabling Movie Club, also disable community
								if (!on && serverConfig.MOVIE_CLUB.communityEnabled) {
									updateMovieClubConfig("communityEnabled", false, () => {
										updateMovieClubConfig("enabled", on, () => {
											movieClubEnabledDisabled = false;
										});
									});
								} else {
									updateMovieClubConfig("enabled", on, () => {
										movieClubEnabledDisabled = false;
									});
								}
							}}
						/>
					</Setting>
					{#if serverConfig.MOVIE_CLUB.enabled}
						<Setting
							title="Enable Movie Club Community"
							desc="Enable community features for movie club including advanced navigation menu."
							row
						>
							<Checkbox
								name="MOVIE_CLUB_COMMUNITY_ENABLED"
								disabled={movieClubCommunityDisabled}
								value={serverConfig.MOVIE_CLUB.communityEnabled}
								toggled={(on) => {
									movieClubCommunityDisabled = true;
									// If disabling community, also disable Matrix chat
									if (!on && serverConfig.MOVIE_CLUB.matrix.enabled) {
										updateMatrixConfig("enabled", false, () => {
											updateMovieClubConfig("communityEnabled", on, () => {
												movieClubCommunityDisabled = false;
											});
										});
									} else {
										updateMovieClubConfig("communityEnabled", on, () => {
											movieClubCommunityDisabled = false;
										});
									}
								}}
							/>
						</Setting>

						{#if serverConfig.MOVIE_CLUB.communityEnabled}
							<Setting
								title="Enable Matrix Chat Integration"
								desc="Enable Matrix/Dendrite integration for community chats during watching phases."
								row
							>
								<Checkbox
									name="MATRIX_ENABLED"
									disabled={matrixEnabledDisabled}
									value={serverConfig.MOVIE_CLUB.matrix.enabled}
									toggled={(on) => {
										matrixEnabledDisabled = true;
										// If disabling Matrix, also disable Application Service
										if (!on && serverConfig.MOVIE_CLUB.matrix.appService.enabled) {
											updateMatrixAppServiceConfig("enabled", false, () => {
												updateMatrixConfig("enabled", on, () => {
													matrixEnabledDisabled = false;
												});
											});
										} else {
											updateMatrixConfig("enabled", on, () => {
												matrixEnabledDisabled = false;
											});
										}
									}}
								/>
							</Setting>

							{#if serverConfig.MOVIE_CLUB.matrix.enabled}
								<Setting
									title="Matrix Server URL"
									desc="URL of your Matrix server (e.g., https://matrix.example.com)"
								>
									<input
										type="url"
										placeholder="https://matrix.example.com"
										bind:value={serverConfig.MOVIE_CLUB.matrix.serverUrl}
										onblur={() => {
											matrixServerUrlDisabled = true;
											updateMatrixConfig(
												"serverUrl",
												serverConfig.MOVIE_CLUB.matrix.serverUrl,
												() => {
													matrixServerUrlDisabled = false;
												},
											);
										}}
										disabled={matrixServerUrlDisabled}
									/>
								</Setting>

								<Setting
									title="Matrix Server Type"
									desc="Type of Matrix server (Auto-detect recommended)"
								>
									<select
										bind:value={serverConfig.MOVIE_CLUB.matrix.serverType}
										onchange={() => {
											matrixServerTypeDisabled = true;
											updateMatrixConfig(
												"serverType",
												serverConfig.MOVIE_CLUB.matrix.serverType,
												() => {
													matrixServerTypeDisabled = false;
												},
											);
										}}
										disabled={matrixServerTypeDisabled}
									>
										<option value="auto">Auto-detect</option>
										<option value="synapse">Synapse</option>
										<option value="dendrite">Dendrite</option>
									</select>
								</Setting>

								<Setting
									title="Matrix Server Name"
									desc="Matrix server domain name (e.g., example.com)"
								>
									<input
										type="text"
										placeholder="example.com"
										bind:value={serverConfig.MOVIE_CLUB.matrix.serverName}
										onblur={() => {
											matrixServerNameDisabled = true;
											updateMatrixConfig(
												"serverName",
												serverConfig.MOVIE_CLUB.matrix.serverName,
												() => {
													matrixServerNameDisabled = false;
												},
											);
										}}
										disabled={matrixServerNameDisabled}
									/>
								</Setting>

								<Setting
									title="Admin Access Token"
									desc="Matrix admin access token for managing users and rooms"
								>
									<input
										type="password"
										placeholder="Enter admin token"
										bind:value={serverConfig.MOVIE_CLUB.matrix.adminToken}
										onblur={() => {
											matrixAdminTokenDisabled = true;
											updateMatrixConfig(
												"adminToken",
												serverConfig.MOVIE_CLUB.matrix.adminToken,
												() => {
													matrixAdminTokenDisabled = false;
												},
											);
										}}
										disabled={matrixAdminTokenDisabled}
									/>
								</Setting>

								<Setting
									title="Admin User ID"
									desc="Matrix user ID for Watcharr admin (e.g., @watcharr:example.com)"
								>
									<input
										type="text"
										placeholder="@watcharr:example.com"
										bind:value={serverConfig.MOVIE_CLUB.matrix.adminUserId}
										onblur={() => {
											matrixAdminUserIdDisabled = true;
											updateMatrixConfig(
												"adminUserId",
												serverConfig.MOVIE_CLUB.matrix.adminUserId,
												() => {
													matrixAdminUserIdDisabled = false;
												},
											);
										}}
										disabled={matrixAdminUserIdDisabled}
									/>
								</Setting>

								<Setting
									title="Space Name"
									desc="Name for the Movie Club space in Matrix"
								>
									<input
										type="text"
										placeholder="Movie Club"
										bind:value={serverConfig.MOVIE_CLUB.matrix.spaceName}
										onblur={() => {
											matrixSpaceNameDisabled = true;
											updateMatrixConfig(
												"spaceName",
												serverConfig.MOVIE_CLUB.matrix.spaceName,
												() => {
													matrixSpaceNameDisabled = false;
												},
											);
										}}
										disabled={matrixSpaceNameDisabled}
									/>
								</Setting>

								<!-- Application Service Configuration -->
								<Setting
									title="Enable Application Service"
									desc="Enable Application Service for virtual Matrix users. Recommended for most installations."
								>
									<Checkbox
										name="APP_SERVICE_ENABLED"
										value={serverConfig.MOVIE_CLUB.matrix.appService.enabled}
										toggled={(on) => {
											updateMatrixAppServiceConfig(
												"enabled",
												on,
												() => {},
											);
										}}
									/>
								</Setting>

								{#if serverConfig.MOVIE_CLUB.matrix.appService.enabled}
									<Setting
										title="Application Service ID"
										desc="Unique identifier for the Application Service (e.g., watcharr-movieclub)"
									>
										<input
											type="text"
											placeholder="Enter AS ID"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.id}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"id",
													serverConfig.MOVIE_CLUB.matrix.appService.id,
													() => {},
												);
											}}
										/>
									</Setting>

									<Setting
										title="AS Token"
										desc="Authentication token for Application Service to homeserver communication"
									>
										<input
											type="password"
											placeholder="AS Token (auto-generated if empty)"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.appServiceToken}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"appServiceToken",
													serverConfig.MOVIE_CLUB.matrix.appService.appServiceToken,
													() => {},
												);
											}}
										/>
									</Setting>

									<Setting
										title="Homeserver Token"
										desc="Authentication token for homeserver to Application Service communication"
									>
										<input
											type="password"
											placeholder="HS Token (auto-generated if empty)"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.homeServerToken}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"homeServerToken",
													serverConfig.MOVIE_CLUB.matrix.appService.homeServerToken,
													() => {},
												);
											}}
										/>
									</Setting>

									<Setting
										title="Sender Localpart"
										desc="Bot user localpart for AS communications (e.g., watcharr-bot)"
									>
										<input
											type="text"
											placeholder="Enter sender localpart"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.senderLocalpart}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"senderLocalpart",
													serverConfig.MOVIE_CLUB.matrix.appService.senderLocalpart,
													() => {},
												);
											}}
										/>
									</Setting>

									<Setting
										title="User Namespace"
										desc="Namespace pattern for AS-managed users (e.g., @watcharr_*:yourdomain.com)"
									>
										<input
											type="text"
											placeholder="@watcharr_*:yourdomain.com"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.userNamespace}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"userNamespace",
													serverConfig.MOVIE_CLUB.matrix.appService.userNamespace,
													() => {},
												);
											}}
										/>
									</Setting>

									<Setting
										title="Alias Namespace"
										desc="Namespace pattern for AS-managed room aliases (e.g., #watcharr_*:yourdomain.com)"
									>
										<input
											type="text"
											placeholder="#watcharr_*:yourdomain.com"
											bind:value={serverConfig.MOVIE_CLUB.matrix.appService.aliasNamespace}
											onblur={() => {
												updateMatrixAppServiceConfig(
													"aliasNamespace",
													serverConfig.MOVIE_CLUB.matrix.appService.aliasNamespace,
													() => {},
												);
											}}
										/>
									</Setting>
								{/if}

								<!-- Registration Secret removed for simplicity - Matrix integration now uses
								     Application Service for virtual users and manual linking for real accounts -->

								<div
									style="display: flex; gap: 10px; flex-wrap: wrap; align-items: stretch;"
								>
									<SettingButton
										title="Test Matrix Connection"
										desc="Quick connection test to Matrix server"
										onClick={() => testMatrixConnection()}
									/>

									<SettingButton
										title="Matrix Setup Validation"
										desc="Comprehensive validation of Matrix configuration and capabilities"
										onClick={() => {
											matrixValidationModalOpen = true;
										}}
									/>

									<SettingButton
										title="Download Registration File"
										desc="Generate and download the Application Service registration file for your Matrix server"
										onClick={() => downloadRegistrationFile()}
									/>

									<SettingButton
										title="Create Rooms for Existing Cycles"
										desc="Create Matrix rooms for active watching cycles that don't have them yet"
										onClick={() => createRoomsForExistingCycles()}
										loading={createRoomsLoading}
									/>
								</div>
							{/if}
						{/if}
					{/if}
					<Setting
						title="Nominations Per User"
						desc="Number of movies each user can nominate per cycle."
					>
						<input
							type="number"
							min="1"
							max="10"
							bind:value={serverConfig.MOVIE_CLUB.nominationsPerUser}
							onblur={() => {
								movieClubNominationsDisabled = true;
								updateMovieClubConfig(
									"nominationsPerUser",
									serverConfig!.MOVIE_CLUB.nominationsPerUser,
									() => {
										movieClubNominationsDisabled = false;
									},
								);
							}}
							disabled={movieClubNominationsDisabled}
						/>
					</Setting>
					<Setting
						title="Votes Per User"
						desc="Number of votes each user can cast per cycle."
					>
						<input
							type="number"
							min="1"
							max="10"
							bind:value={serverConfig.MOVIE_CLUB.votesPerUser}
							onblur={() => {
								movieClubVotesDisabled = true;
								updateMovieClubConfig(
									"votesPerUser",
									serverConfig!.MOVIE_CLUB.votesPerUser,
									() => {
										movieClubVotesDisabled = false;
									},
								);
							}}
							disabled={movieClubVotesDisabled}
						/>
					</Setting>
					<Setting
						title="Phase Duration (Days)"
						desc="How long each phase (nomination, voting, watching) lasts in days."
					>
						<input
							type="number"
							min="1"
							max="30"
							bind:value={serverConfig.MOVIE_CLUB.phaseDurationDays}
							onblur={() => {
								movieClubDurationDisabled = true;
								updateMovieClubConfig(
									"phaseDurationDays",
									serverConfig!.MOVIE_CLUB.phaseDurationDays,
									() => {
										movieClubDurationDisabled = false;
									},
								);
							}}
							disabled={movieClubDurationDisabled}
						/>
					</Setting>

					<div>
						<h3>Services</h3>
						<h5 class="norm">
							These integrations are not in their final stages. Consider them a
							preview/beta, if you have any issues,
							<a
								style="text-decoration: underline;"
								href="https://github.com/sbondCo/Watcharr/issues/new/choose"
								target="_blank"
							>
								please report them.
							</a>
						</h5>
					</div>

					<Setting title="Twitch">
						<SettingButton
							title="Twitch"
							desc="Twitch application credentials for enabling game support (via IGDB)."
							icon={Object.keys(serverConfig.TWITCH).length > 0
								? "arrow"
								: "add"}
							onClick={() => {
								twitchModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Sonarr">
						{#if serverConfig.SONARR?.length > 0}
							{#each serverConfig.SONARR as server}
								<SettingButton
									title={server.name}
									desc={`Configure server at ${server.host}`}
									onClick={() => {
										sonarrServerEditing = server;
										sonarrModalEditing = true;
										sonarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Sonarr"
							desc="Add a Sonarr server."
							icon="add"
							onClick={() => {
								let name = "Sonarr";
								if (serverConfig!.SONARR?.length > 0) {
									// if this still exists ya on yur own
									name = `Sonarr${serverConfig!.SONARR.length + 1}`;
								}
								sonarrServerEditing = { name };
								sonarrModalEditing = false;
								sonarrModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Radarr">
						{#if serverConfig.RADARR?.length > 0}
							{#each serverConfig.RADARR as server}
								<SettingButton
									title={server.name}
									desc={`Configure server at ${server.host}`}
									onClick={() => {
										radarrServerEditing = server;
										radarrModalEditing = true;
										radarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Radarr"
							desc="Add a Radarr server."
							icon="add"
							onClick={() => {
								let name = "Radarr";
								if (serverConfig!.RADARR?.length > 0) {
									// if this still exists ya on yur own
									name = `Radarr${serverConfig!.RADARR.length + 1}`;
								}
								radarrServerEditing = { name };
								radarrModalEditing = false;
								radarrModalOpen = true;
							}}
						/>
					</Setting>

					{#if twitchModalOpen}
						<TwitchModal
							cfg={serverConfig.TWITCH}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								twitchModalOpen = false;
							}}
						/>
					{/if}

					{#if sonarrModalOpen && sonarrServerEditing}
						<SonarrModal
							servarr={sonarrServerEditing}
							isEditing={sonarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								sonarrModalOpen = false;
								sonarrServerEditing = undefined;
							}}
						/>
					{/if}

					{#if radarrModalOpen && radarrServerEditing}
						<RadarrModal
							servarr={radarrServerEditing}
							isEditing={radarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								radarrModalOpen = false;
								radarrServerEditing = undefined;
							}}
						/>
					{/if}

					{#if matrixValidationModalOpen}
						<MatrixValidation
							onClose={() => {
								matrixValidationModalOpen = false;
							}}
						/>
					{/if}
				{/if}
			{:catch err}
				<PageError error={err} pretty="Failed to load server config" />
			{/await}
		</SettingsList>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			min-width: 400px;
			max-width: 400px;
			overflow: hidden;

			h2 {
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
			}

			& > div:not(:first-of-type) {
				margin-top: 30px;
			}

			@media screen and (max-width: 440px) {
				width: 100%;
				min-width: unset;
			}
		}
	}
</style>
