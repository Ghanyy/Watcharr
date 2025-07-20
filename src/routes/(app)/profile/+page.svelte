<script lang="ts">
	import { goto } from "$app/navigation";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import Stats from "@/lib/stats/Stats.svelte";
	import { updateUserSetting } from "@/lib/util/api";
	import { getOrdinalSuffix, monthsShort } from "@/lib/util/helpers";
	import { store } from "@/store.svelte";
	import {
		UserType,
		type Image,
		type Profile,
		type MatrixAccountInfo,
	} from "@/types";
	import axios from "axios";
	import { notify } from "@/lib/util/notify";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import PwChangeModal from "@/routes/(app)/profile/modals/PwChangeModal.svelte";
	import SyncModal from "./modals/SyncModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import RatingSetting from "@/lib/rating/RatingSetting.svelte";
	import { toggleTheme } from "@/lib/util/theme";

	let user = $derived(store.userInfo);
	let settings = $derived(store.userSettings);
	let selectedTheme = $derived(store.appTheme);
	let config = $derived(store.config);

	let privateDisabled = $state(false);
	let privateThoughtsDisabled = $state(false);
	let exportDisabled = $state(false);
	let hideSpoilersDisabled = $state(false);
	let countryDisabled = $state(false);
	let includePreviouslyWatchedDisabled = $state(false);
	let automateShowStatusesDisabled = $state(false);
	let pwChangeModalOpen = $state(false);
	let getProfilePromise = $state(getProfile());
	let jellyfinSyncModalOpen = $state(false);
	let plexSyncModalOpen = $state(false);

	// Matrix state variables
	let matrixInfo = $state<MatrixAccountInfo | null>(null);
	let matrixInfoLoading = $state(false);
	let matrixCreateLoading = $state(false);
	let matrixLinkLoading = $state(false);
	let matrixUnlinkLoading = $state(false);
	let customMatrixUserId = $state("");
	let customAccessToken = $state("");
	let showLinkForm = $state(false);
	let exportCredentialsLoading = $state(false);
	let showExportModal = $state(false);
	let exportedCredentials = $state<{
		matrixUserId: string;
		password: string;
		serverUrl: string;
		serverName: string;
		exportedAt: string;
	} | null>(null);

	async function getProfile() {
		return (await axios.get(`/profile`)).data as Profile;
	}

	function formatDate(d: Date) {
		return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${
			monthsShort[d.getMonth()]
		} ${d.getFullYear()}`;
	}

	function updateBio(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLTextAreaElement },
	) {
		const newBio = ev?.currentTarget?.value;
		if (typeof newBio !== "string") {
			console.warn("updateBio called without any value", newBio);
			return;
		}
		const nid = notify({ text: "Updating Bio", type: "loading" });
		axios
			.post("/user/bio", { newBio: newBio })
			.then(() => {
				if (user) {
					user.bio = newBio;
					notify({ id: nid, text: "Updated Bio", type: "success" });
				}
			})
			.catch((err) => {
				notify({
					id: nid,
					text: err?.response?.data?.error ?? "Failed to update bio",
					type: "error",
				});
			});
	}

	function avatarDropped(ev: Event) {
		const files = (ev.currentTarget as HTMLInputElement)?.files;
		if (!files || files?.length <= 0) {
			console.error("avatarDropped: no file found");
			return;
		}
		const nid = notify({ text: "Uploading avatar", type: "loading" });
		axios
			.postForm(
				"/user/avatar",
				{ avatar: files[0] },
				{
					headers: {
						"Content-Type": "multipart/form-data",
					},
				},
			)
			.then((r) => {
				if (user) {
					user.avatar = r.data as Image;
					notify({ id: nid, text: "Avatar Uploaded", type: "success" });
				}
			})
			.catch((err) => {
				console.error("uploading avatar failed", err);
				notify({
					id: nid,
					text: err?.response?.data?.error ?? "Failed to upload avatar",
					type: "error",
				});
			});
	}

	async function downloadWatchedList() {
		const nid = notify({ text: "Exporting", type: "loading" });
		try {
			exportDisabled = true;
			// We re-fetch, to ensure data we export is up to date.
			const r = await axios.get("/watched");
			console.log(r.data);
			if (!r.data || r.data?.length <= 0) {
				notify({
					id: nid,
					text: "Can't export an empty watch list!",
					type: "error",
					time: 10000,
				});
				exportDisabled = false;
				return;
			}
			const file = new Blob([JSON.stringify(r.data, undefined, 2)], {
				type: "application/json",
			});
			const a = document.createElement("a");
			a.href = URL.createObjectURL(file);
			a.download = "watcharr-export.json";
			a.click();
			exportDisabled = false;
			notify({ id: nid, text: "Successfully Exported", type: "success" });
		} catch (err) {
			console.error("downloadWatchedList failed!", err);
			notify({ id: nid, text: "Export Failed!", type: "error" });
		}
	}

	/**
	 * Takes in number of minutes and converts to readable.
	 * eg into months, weeks, days, hours and minutes.
	 */
	function toFormattedTimeLong(m: number) {
		// Considers a 30 days long month
		const countInMinutes = [
			["month", 43200],
			["week", 10080],
			["day", 1440],
			["hour", 60],
		];
		let ansString = "";
		let tmp;
		for (const c of countInMinutes) {
			tmp = Math.floor(m / (c[1] as number));

			// Ignore fields with fewer than 1 unit
			if (tmp) ansString += `${tmp} ${c[0]}${tmp >= 2 ? "s, " : ", "}`;
			m -= tmp * (c[1] as number);
		}
		if (!ansString) {
			return "0 hours";
		}
		return ansString.slice(0, -2);
	}

	// Matrix functions
	async function getMatrixInfo() {
		if (!config?.MOVIE_CLUB?.matrix?.enabled) return;

		matrixInfoLoading = true;
		try {
			const response = await axios.get("/matrix/info");
			matrixInfo = response.data;
		} catch (error) {
			console.error("Failed to get Matrix info:", error);
			notify({ type: "error", text: "Failed to load Matrix account info" });
		} finally {
			matrixInfoLoading = false;
		}
	}

	async function createMatrixAccount() {
		matrixCreateLoading = true;
		try {
			const response = await axios.post("/matrix/create-account");
			if (response.status === 200) {
				notify({
					type: "success",
					text: "Matrix account created successfully",
				});
				getMatrixInfo(); // Refresh info
			}
		} catch (error) {
			console.error("Failed to create Matrix account:", error);
			notify({ type: "error", text: "Failed to create Matrix account" });
		} finally {
			matrixCreateLoading = false;
		}
	}

	async function linkCustomMatrixAccount() {
		if (!customMatrixUserId || !customAccessToken) {
			notify({
				type: "error",
				text: "Please provide both Matrix User ID and Access Token",
			});
			return;
		}

		matrixLinkLoading = true;
		try {
			const response = await axios.post("/matrix/link-account", {
				matrixUserId: customMatrixUserId,
				accessToken: customAccessToken,
			});
			if (response.status === 200) {
				notify({ type: "success", text: "Matrix account linked successfully" });
				customMatrixUserId = "";
				customAccessToken = "";
				getMatrixInfo(); // Refresh info
			}
		} catch (error) {
			console.error("Failed to link Matrix account:", error);
			notify({ type: "error", text: "Failed to link Matrix account" });
		} finally {
			matrixLinkLoading = false;
		}
	}

	async function unlinkMatrixAccount() {
		matrixUnlinkLoading = true;
		try {
			const response = await axios.delete("/matrix/unlink-account");
			if (response.status === 200) {
				notify({
					type: "success",
					text: "Matrix account unlinked successfully",
				});
				getMatrixInfo(); // Refresh info
				showLinkForm = false;
				customMatrixUserId = "";
				customAccessToken = "";
			}
		} catch (error) {
			console.error("Failed to unlink Matrix account:", error);
			notify({ type: "error", text: "Failed to unlink Matrix account" });
		} finally {
			matrixUnlinkLoading = false;
		}
	}

	async function exportMatrixCredentials(deletePasswordAfterExport = false) {
		exportCredentialsLoading = true;
		try {
			const response = await axios.post("/matrix/export-credentials", {
				deletePasswordAfterExport,
			});

			if (response.data?.credentials) {
				exportedCredentials = response.data.credentials;
				showExportModal = true;
				notify({
					type: "success",
					text: "Matrix credentials exported successfully",
				});
			}
		} catch (error: any) {
			console.error("Failed to export Matrix credentials:", error);
			let errorMessage = "Failed to export Matrix credentials";

			if (error.response?.status === 404) {
				errorMessage = "No Matrix account found";
			} else if (error.response?.status === 403) {
				errorMessage =
					"Credential export is only available for auto-generated accounts";
			} else if (error.response?.status === 429) {
				errorMessage =
					error.response?.data?.error ||
					"Rate limit exceeded - please wait before trying again";
			} else if (error.response?.data?.error) {
				errorMessage = error.response.data.error;
			}

			notify({ type: "error", text: errorMessage });
		} finally {
			exportCredentialsLoading = false;
		}
	}

	function toggleLinkForm() {
		showLinkForm = !showLinkForm;
		if (!showLinkForm) {
			customMatrixUserId = "";
			customAccessToken = "";
		}
	}

	function closeExportModal() {
		showExportModal = false;
		exportedCredentials = null;
	}

	function copyToClipboard(text: string, label: string) {
		navigator.clipboard
			.writeText(text)
			.then(() => {
				notify({ type: "success", text: `${label} copied to clipboard` });
			})
			.catch(() => {
				notify({ type: "error", text: `Failed to copy ${label}` });
			});
	}

	// Helper function to get account type display information
	function getAccountTypeInfo(
		accountType?: string,
		isAutoGenerated?: boolean,
		asManagedUser?: boolean,
	) {
		if (accountType === "appservice") {
			return {
				type: "Application Service",
				description: "Virtual Matrix user managed by Watcharr",
				icon: "🤖",
				color: "#4A90E2",
				features: [
					"Movie Club Chats",
					"No External Access",
					"Automatic Management",
				],
			};
		} else if (accountType === "personal") {
			return {
				type: isAutoGenerated
					? "Personal (Auto-Generated)"
					: "Personal (Custom Linked)",
				description: isAutoGenerated
					? "Real Matrix account created by Watcharr"
					: "Your existing Matrix account linked to Watcharr",
				icon: "👤",
				color: "#22C55E",
				features: isAutoGenerated
					? ["Movie Club Chats", "Element Web Access", "Credential Export"]
					: ["Movie Club Chats", "Element Web Access", "Your Existing Account"],
			};
		} else if (accountType === "legacy") {
			return {
				type: "Legacy Account",
				description: "Older account format that should be migrated",
				icon: "⚠️",
				color: "#F59E0B",
				features: ["Limited Features", "Migration Recommended"],
			};
		} else {
			return {
				type: "Unknown",
				description: isAutoGenerated
					? "Auto-generated account"
					: "Custom linked account",
				icon: "❓",
				color: "#6B7280",
				features: ["Basic Access"],
			};
		}
	}

	// Load Matrix info when component loads
	$effect(() => {
		if (config?.MOVIE_CLUB?.matrix?.enabled) {
			getMatrixInfo();
		}
	});
</script>

<svelte:head>
	<title>My Profile</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="user-basic-info">
			<UserAvatar img={user?.avatar} {avatarDropped} />
			<div>
				<h2 title={user?.username}>
					<span style="font-weight: normal; font-variant: all-small-caps;"
						>Hey</span
					>
					{user?.username}
				</h2>
				<textarea
					rows="1"
					placeholder="my bio"
					onblur={updateBio}
					value={user?.bio}
				></textarea>
			</div>
		</div>

		<Stats>
			{#await getProfilePromise}
				<Spinner />
			{:then profile}
				<Stat name="Joined" value={formatDate(new Date(profile.joined))} />
				<Stat name="Movies Watched" value={profile.moviesWatched} large />
				<Stat name="Shows Watched" value={profile.showsWatched} large />
				<Stat
					name="Watching Movies"
					value={toFormattedTimeLong(profile.moviesWatchedRuntime)}
				/>
				<Stat
					name="Watching Shows"
					value={toFormattedTimeLong(profile.showsWatchedRuntime)}
					disc="This is very inaccurate 🚀"
				/>
			{:catch err}
				<Error error={err} pretty="Failed to get stats!" />
			{/await}
		</Stats>

		<div class="settings">
			<h3 class="norm">Settings</h3>

			<div class="theme">
				<h4 class="norm">Theme</h4>
				<div class="row">
					<button
						class={`plain${selectedTheme === "system" ? " selected" : ""}`}
						id="system"
						onclick={() => toggleTheme("system")}
					>
						<span>system</span>
					</button>
					<button
						class={`plain${selectedTheme === "light" ? " selected" : ""}`}
						id="light"
						onclick={() => toggleTheme("light")}
					>
						light
					</button>
					<button
						class={`plain${selectedTheme === "dark" ? " selected" : ""}`}
						id="dark"
						onclick={() => toggleTheme("dark")}
					>
						dark
					</button>
				</div>
			</div>

			<Setting
				title="Country"
				desc="What country would you like to see available streaming providers for?"
			>
				<RegionDropDown
					selectedCountry={settings?.country}
					disabled={countryDisabled}
					onChange={(c) => {
						countryDisabled = true;
						updateUserSetting("country", c, () => {
							countryDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting title="Private" desc="Hide your profile from others?" row>
				<Checkbox
					name="private"
					disabled={privateDisabled}
					value={settings?.private}
					toggled={(on) => {
						privateDisabled = true;
						updateUserSetting("private", on, () => {
							privateDisabled = false;
						});
					}}
				/>
			</Setting>

			{#if !settings?.private}
				<Setting
					title="Private Thoughts"
					desc="Hide your watched list thoughts from followers?"
					row
				>
					<Checkbox
						name="privateThoughts"
						disabled={privateDisabled}
						value={settings?.privateThoughts}
						toggled={(on) => {
							privateThoughtsDisabled = true;
							updateUserSetting("privateThoughts", on, () => {
								privateThoughtsDisabled = false;
							});
						}}
					/>
				</Setting>
			{/if}

			<Setting
				title="Hide Spoilers"
				desc="Do you want to hide episode info?"
				row
			>
				<Checkbox
					name="hideSpoilers"
					disabled={hideSpoilersDisabled}
					value={settings?.hideSpoilers}
					toggled={(on) => {
						hideSpoilersDisabled = true;
						updateUserSetting("hideSpoilers", on, () => {
							hideSpoilersDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title="Automate Show Statuses"
				desc="Do you want to automate show statuses (show, season, episode)?"
				tag="experimental"
				row
			>
				<Checkbox
					name="automateShowStatusesDisabled"
					disabled={automateShowStatusesDisabled}
					value={settings?.automateShowStatuses}
					toggled={(on) => {
						automateShowStatusesDisabled = true;
						updateUserSetting("automateShowStatuses", on, () => {
							automateShowStatusesDisabled = false;
						});
					}}
				/>
			</Setting>

			<Setting
				title="Include Previously Watched"
				desc="Do you want to include previously watched content in the 'finished' filter, even if their status has now been changed?"
				row
			>
				<Checkbox
					name="includePreviouslyWatched"
					disabled={includePreviouslyWatchedDisabled}
					value={settings?.includePreviouslyWatched}
					toggled={(on) => {
						includePreviouslyWatchedDisabled = true;
						updateUserSetting("includePreviouslyWatched", on, () => {
							includePreviouslyWatchedDisabled = false;
							// Get profile stats again
							getProfilePromise = getProfile();
						});
					}}
				/>
			</Setting>

			<RatingSetting />

			{#if config?.MOVIE_CLUB?.matrix?.enabled}
				<h4 class="norm">Matrix Chat Account</h4>
				{#if matrixInfoLoading}
					<div style="display: flex; justify-content: center; padding: 20px;">
						<Spinner />
					</div>
				{:else if matrixInfo}
					{#if matrixInfo.hasMatrixAccount}
						{@const accountTypeInfo = getAccountTypeInfo(
							matrixInfo.accountType,
							matrixInfo.isAutoGenerated,
							matrixInfo.asManagedUser,
						)}
						<Setting
							title="Matrix Account"
							desc="Your Matrix account for Movie Club community chats"
						>
							<div style="display: flex; flex-direction: column; gap: 15px;">
								<!-- Account Type Badge -->
								<div
									class="matrix-account-type"
									style="border-left: 3px solid {accountTypeInfo.color};"
								>
									<div class="account-type-header">
										<span class="account-type-icon">{accountTypeInfo.icon}</span
										>
										<div class="account-type-info">
											<span class="account-type-name"
												>{accountTypeInfo.type}</span
											>
											<span class="account-type-desc"
												>{accountTypeInfo.description}</span
											>
										</div>
									</div>
									<div class="account-features">
										{#each accountTypeInfo.features as feature}
											<span class="feature-tag">{feature}</span>
										{/each}
									</div>
								</div>

								<!-- User ID Display -->
								<div style="display: flex; flex-direction: column; gap: 5px;">
									<span
										style="font-weight: 500; font-family: monospace; font-size: 14px;"
										>{matrixInfo.matrixUserId}</span
									>
									{#if matrixInfo.accountType === "appservice"}
										<span
											style="font-size: 12px; opacity: 0.7; color: {accountTypeInfo.color};"
										>
											Virtual user - managed automatically by Watcharr
										</span>
									{:else if matrixInfo.accountType === "personal"}
										<span
											style="font-size: 12px; opacity: 0.7; color: {accountTypeInfo.color};"
										>
											{matrixInfo.isAutoGenerated
												? "Real Matrix account - can be used with Element Web"
												: "Your linked Matrix account"}
										</span>
									{:else if matrixInfo.accountType === "legacy"}
										<span
											style="font-size: 12px; opacity: 0.7; color: {accountTypeInfo.color};"
										>
											⚠️ Legacy account format - consider recreating for latest
											features
										</span>
									{:else}
										<span style="font-size: 12px; opacity: 0.7;">
											{matrixInfo.isAutoGenerated
												? "Auto-generated account"
												: "Custom linked account"}
										</span>
									{/if}
								</div>

								<div style="display: flex; gap: 10px; flex-wrap: wrap;">
									{#if matrixInfo.accountType !== "appservice"}
										<button
											onclick={toggleLinkForm}
											style="align-self: flex-start;"
										>
											{showLinkForm ? "Cancel" : "Change Account"}
										</button>
									{/if}
									{#if matrixInfo.accountType === "personal" && matrixInfo.isAutoGenerated}
										<button
											onclick={() => exportMatrixCredentials(false)}
											disabled={exportCredentialsLoading}
											style="align-self: flex-start; background-color: #4CAF50; color: white;"
										>
											{#if exportCredentialsLoading}
												<Spinner />
											{:else}
												Export Credentials
											{/if}
										</button>
									{/if}
									<button
										onclick={unlinkMatrixAccount}
										disabled={matrixUnlinkLoading}
										style="align-self: flex-start; background-color: #f44336; color: white;"
									>
										{#if matrixUnlinkLoading}
											<Spinner />
										{:else}
											{matrixInfo.accountType === "appservice"
												? "Remove Account"
												: "Unlink Account"}
										{/if}
									</button>
								</div>

								{#if showLinkForm}
									<div
										style="border-top: 1px solid var(--border-color); padding-top: 15px;"
									>
										<h5 style="margin: 0 0 10px 0; font-size: 14px;">
											Link New Account
										</h5>
										<div
											style="display: flex; flex-direction: column; gap: 10px;"
										>
											<input
												type="text"
												placeholder="@username:matrix.server.com"
												bind:value={customMatrixUserId}
												disabled={matrixLinkLoading}
											/>
											<input
												type="password"
												placeholder="Access Token"
												bind:value={customAccessToken}
												disabled={matrixLinkLoading}
											/>
											<button
												onclick={linkCustomMatrixAccount}
												disabled={matrixLinkLoading ||
													!customMatrixUserId ||
													!customAccessToken}
												style="align-self: flex-start;"
											>
												{#if matrixLinkLoading}
													<Spinner />
												{:else}
													Link Account
												{/if}
											</button>
										</div>
									</div>
								{/if}
							</div>
						</Setting>
					{:else}
						<Setting
							title="Matrix Account"
							desc="Set up a Matrix account to access Movie Club community chats"
						>
							<div style="display: flex; flex-direction: column; gap: 20px;">
								<!-- Account Type Options -->
								<div class="matrix-account-options">
									{#if config?.MOVIE_CLUB?.matrix?.appService?.enabled}
										<div class="account-option recommended">
											<div class="option-header">
												<span class="option-icon">🤖</span>
												<div class="option-info">
													<span class="option-name"
														>Application Service Account</span
													>
													<span class="option-badge">Recommended</span>
												</div>
											</div>
											<p class="option-description">
												Virtual Matrix user managed automatically by Watcharr.
												Perfect for users who only need Movie Club features.
											</p>
											<div class="option-features">
												<span class="feature-tag">✓ Automatic Setup</span>
												<span class="feature-tag">✓ No Credentials Needed</span>
												<span class="feature-tag">✓ Movie Club Chats</span>
											</div>
											<button
												onclick={createMatrixAccount}
												disabled={matrixCreateLoading}
												class="option-button primary"
											>
												{#if matrixCreateLoading}
													<Spinner />
												{:else}
													Create Account
												{/if}
											</button>
										</div>
									{:else}
										<div class="account-option disabled">
											<div class="option-header">
												<span class="option-icon">🤖</span>
												<div class="option-info">
													<span class="option-name"
														>Application Service Account</span
													>
													<span class="option-badge disabled">Not Available</span>
												</div>
											</div>
											<p class="option-description">
												Application Service is not enabled on this server. Contact your administrator to enable AS features.
											</p>
											<div class="option-features">
												<span class="feature-tag disabled">✗ AS Not Enabled</span>
												<span class="feature-tag disabled">✗ Contact Admin</span>
											</div>
											<button
												disabled
												class="option-button primary disabled"
											>
												AS Not Available
											</button>
										</div>
									{/if}

									<div class="account-option">
										<div class="option-header">
											<span class="option-icon">👤</span>
											<div class="option-info">
												<span class="option-name">Link Existing Matrix Account</span>
											</div>
										</div>
										<p class="option-description">
											Connect your existing Matrix account that you can access with Element Web
											and other Matrix clients.
										</p>
										<div class="option-features">
											<span class="feature-tag">✓ Element Web Access</span>
											<span class="feature-tag">✓ Full Matrix Features</span>
											<span class="feature-tag">✓ Your Existing Account</span>
										</div>
										<div class="sub-option">
											<div class="link-form-inline">
												<input
													type="text"
													placeholder="@username:matrix.server.com"
													bind:value={customMatrixUserId}
													disabled={matrixLinkLoading}
												/>
												<input
													type="password"
													placeholder="Access Token"
													bind:value={customAccessToken}
													disabled={matrixLinkLoading}
												/>
												<button
													onclick={linkCustomMatrixAccount}
													disabled={matrixLinkLoading ||
														!customMatrixUserId ||
														!customAccessToken}
													class="option-button secondary"
												>
													{#if matrixLinkLoading}
														<Spinner />
													{:else}
														Link Account
													{/if}
												</button>
											</div>
										</div>
									</div>
								</div>
							</div>
						</Setting>
					{/if}
				{/if}
			{/if}

			<div class="row btns">
				<button onclick={() => goto("/import")}>Import</button>
				<button onclick={() => downloadWatchedList()} disabled={exportDisabled}
					>Export</button
				>
				{#if user?.type !== UserType.Plex && user?.type !== UserType.Jellyfin}
					<button
						onclick={() => {
							pwChangeModalOpen = true;
						}}>Change Password</button
					>
				{/if}
				{#if user?.type === UserType?.Jellyfin}
					<button
						onclick={() => (jellyfinSyncModalOpen = true)}
						disabled={exportDisabled}
					>
						Sync With {localStorage.getItem("useEmby") ? "Emby" : "Jellyfin"}
					</button>
				{/if}
				{#if user?.type === UserType?.Plex}
					<button
						onclick={() => (plexSyncModalOpen = true)}
						disabled={exportDisabled}
					>
						Sync With Plex
					</button>
				{/if}
			</div>
			{#if pwChangeModalOpen}
				<PwChangeModal
					userName={user?.username}
					onClose={() => {
						pwChangeModalOpen = false;
					}}
				></PwChangeModal>
			{/if}
			{#if jellyfinSyncModalOpen}
				<SyncModal onClose={() => (jellyfinSyncModalOpen = false)} />
			{/if}
			{#if plexSyncModalOpen}
				<SyncModal type="plex" onClose={() => (plexSyncModalOpen = false)} />
			{/if}
			{#if showExportModal && exportedCredentials}
				<div class="modal-backdrop" onclick={closeExportModal}>
					<div class="credentials-modal" onclick={(e) => e.stopPropagation()}>
						<div class="modal-header">
							<h2>Matrix Credentials Exported</h2>
							<button class="close-button" onclick={closeExportModal}>×</button>
						</div>

						<div class="modal-content">
							<div class="export-intro">
								<p>
									Your Matrix credentials have been exported successfully. Use
									these credentials to access Element Web or other Matrix
									clients.
								</p>
							</div>

							<div class="credentials-section">
								<h4>🔑 Your Credentials</h4>

								<div
									class="credential-item"
									style="border-left: 3px solid var(--accent-color)"
								>
									<div class="credential-header">
										<div class="credential-info">
											<span class="credential-icon">👤</span>
											<span class="credential-name">Matrix User ID</span>
										</div>
										<button
											class="copy-button"
											onclick={() =>
												copyToClipboard(
													exportedCredentials!.matrixUserId,
													"User ID",
												)}
										>
											📋 Copy
										</button>
									</div>
									<div class="credential-value">
										{exportedCredentials.matrixUserId}
									</div>
								</div>

								<div
									class="credential-item"
									style="border-left: 3px solid var(--accent-color)"
								>
									<div class="credential-header">
										<div class="credential-info">
											<span class="credential-icon">🔒</span>
											<span class="credential-name">Password</span>
										</div>
										<button
											class="copy-button"
											onclick={() =>
												copyToClipboard(
													exportedCredentials!.password,
													"Password",
												)}
										>
											📋 Copy
										</button>
									</div>
									<div class="credential-value monospace">
										{exportedCredentials.password}
									</div>
								</div>

								<div
									class="credential-item"
									style="border-left: 3px solid var(--accent-color)"
								>
									<div class="credential-header">
										<div class="credential-info">
											<span class="credential-icon">🌐</span>
											<span class="credential-name">Server URL</span>
										</div>
										<button
											class="copy-button"
											onclick={() =>
												copyToClipboard(
													exportedCredentials!.serverUrl,
													"Server URL",
												)}
										>
											📋 Copy
										</button>
									</div>
									<div class="credential-value">
										{exportedCredentials.serverUrl}
									</div>
								</div>
							</div>

							<div class="setup-guide">
								<h4>📱 Element Web Setup</h4>
								<div class="setup-steps">
									<ol>
										<li>
											Visit <a
												href="https://app.element.io"
												target="_blank"
												rel="noopener">app.element.io</a
											>
										</li>
										<li>Click "Sign In"</li>
										<li>Click "Edit" next to the server field</li>
										<li>
											Enter your server URL: <code
												>{exportedCredentials.serverUrl}</code
											>
										</li>
										<li>Use your Matrix User ID and Password above</li>
									</ol>
								</div>
							</div>

							<div class="security-notice">
								<h4>⚠️ Security Notice</h4>
								<p>
									Store these credentials securely. For enhanced security, you
									can export again with the "Delete password after export"
									option to remove the password from our database.
								</p>
							</div>
						</div>

						<div class="modal-footer">
							<div class="footer-actions">
								<button
									onclick={() => {
										closeExportModal();
										exportMatrixCredentials(true);
									}}
									class="warning"
								>
									🗑️ Export & Delete Password
								</button>
								<button onclick={closeExportModal} class="primary">Close</button
								>
							</div>
						</div>
					</div>
				</div>
			{/if}
		</div>
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

	.user-basic-info {
		display: flex;
		gap: 20px;

		& > div {
			display: flex;
			flex-flow: column;
			gap: 5px;
			width: 100%;
			overflow: hidden;

			textarea {
				resize: none;

				&:not(:focus) {
					border: 0;
					padding: 0;
					height: 32px;
				}
			}
		}
	}

	.settings {
		display: flex;
		flex-flow: column;
		gap: 20px;
		width: 100%;

		h3 {
			font-variant: small-caps;
		}

		& > div {
			margin: 0 15px;
		}

		div {
			&.row {
				display: flex;
				flex-flow: row;
				gap: 10px;
				align-items: center;

				&.btns button {
					width: max-content;
				}
			}
		}

		.theme {
			display: flex;
			flex-flow: column;
			gap: 10px;

			& button {
				width: 50%;
				height: 80px;
				border-radius: 10px;
				outline: 3px solid;
				font-size: 20px;
				text-transform: uppercase;
				font-family: "Rampart One";
				color: transparent;
				transition: all 200ms ease-in;

				&#light {
					background-color: white;
					outline-color: $accent-color;
					&:hover {
						color: black;
						-webkit-text-stroke: 0.5px black;
					}
				}

				&#dark {
					background-color: black;
					outline-color: white;
					&:hover {
						color: white;
						-webkit-text-stroke: 0.5px white;
					}
				}

				&#system {
					background: linear-gradient(to right bottom, white 50%, black 50.3%);
					outline-color: black;

					span {
						mix-blend-mode: difference;
					}

					&:hover {
						color: white;
						-webkit-text-stroke: 0.5px white;
					}
				}

				&.selected {
					outline-color: gold !important;
				}
			}
		}
	}

	// Matrix Credential Export Modal Styles
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-color: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(4px);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
		padding: 20px;
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: 10px;
		width: 90%;
		max-width: 600px;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
		border: 1px solid var(--border-color);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 20px;
		border-bottom: 1px solid var(--border-color);

		h3 {
			margin: 0;
			color: var(--fg);
		}

		.close-btn {
			background: none;
			border: none;
			font-size: 24px;
			cursor: pointer;
			color: var(--fg);
			padding: 0;
			width: 30px;
			height: 30px;
			display: flex;
			align-items: center;
			justify-content: center;

			&:hover {
				background-color: var(--border-color);
				border-radius: 50%;
			}
		}
	}

	.modal-body {
		padding: 20px;

		.credential-section {
			margin-bottom: 20px;

			label {
				display: block;
				font-weight: 500;
				margin-bottom: 5px;
				color: var(--fg);
			}

			.credential-field {
				display: flex;
				gap: 8px;
				align-items: stretch;

				code {
					background-color: var(--bg-secondary);
					padding: 12px 16px;
					border-radius: 6px;
					border: 1px solid var(--border-color);
					flex: 1;
					font-family: "Courier New", monospace;
					word-break: break-all;
					font-size: 13px;
					line-height: 1.4;
					display: flex;
					align-items: center;
					color: var(--fg);
				}

				button {
					padding: 6px 12px;
					background-color: var(--accent);
					color: white;
					border: none;
					border-radius: 6px;
					cursor: pointer;
					font-size: 11px;
					font-weight: 500;
					min-width: 50px;
					flex-shrink: 0;

					&:hover {
						opacity: 0.9;
					}
				}
			}
		}

		.element-setup {
			background-color: var(--bg-primary);
			padding: 15px;
			border-radius: 8px;
			border: 1px solid var(--border-color);
			margin: 20px 0;

			h4 {
				margin: 0 0 10px 0;
				color: var(--fg);
			}

			ol {
				margin: 0;
				padding-left: 20px;

				li {
					margin-bottom: 5px;
					color: var(--fg);

					a {
						color: #2196f3;
						text-decoration: none;

						&:hover {
							text-decoration: underline;
						}
					}

					code {
						background-color: var(--bg-secondary);
						padding: 2px 6px;
						border-radius: 3px;
						font-size: 12px;
					}
				}
			}
		}

		.security-warning {
			background-color: var(--bg-secondary);
			border: 1px solid var(--border-color);
			padding: 15px;
			border-radius: 8px;
			margin-top: 20px;

			h4 {
				margin: 0 0 10px 0;
				color: var(--fg);
			}

			p {
				margin: 0 0 15px 0;
				color: var(--fg);
				font-size: 14px;
				opacity: 0.9;
			}

			.delete-password-btn {
				background-color: #ff9800;
				color: white;
				border: none;
				padding: 8px 16px;
				border-radius: 6px;
				cursor: pointer;
				font-size: 13px;
				font-weight: 500;

				&:hover {
					background-color: #f57c00;
				}
			}
		}
	}

	// Matrix Credential Export Modal Styles (based on MatrixValidation.svelte)
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

	.credentials-modal {
		background: var(--bg-color);
		border-radius: 12px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
		width: 90%;
		max-width: 700px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.credentials-modal .modal-header {
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

	.credentials-modal .modal-content {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}

	.export-intro {
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

	.credentials-section {
		display: flex;
		flex-direction: column;
		gap: 15px;
		margin-bottom: 25px;

		h4 {
			margin: 0 0 15px 0;
			font-size: 1.1em;
			color: var(--text-color);
		}
	}

	.credential-item {
		background: var(--bg-color);
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid var(--accent-color);
		margin-bottom: 15px;
	}

	.credential-header {
		padding: 15px 15px 10px 15px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 15px;

		.credential-info {
			display: flex;
			align-items: center;
			gap: 10px;
			flex: 1;

			.credential-icon {
				font-size: 1.2em;
			}

			.credential-name {
				font-weight: 500;
				font-size: 1.05em;
			}
		}

		.copy-button {
			background: var(--accent-color);
			color: white;
			border: none;
			padding: 8px 12px;
			border-radius: 6px;
			cursor: pointer;
			font-size: 0.85em;
			font-weight: 500;
			transition: all 0.2s ease;
			white-space: nowrap;
			flex-shrink: 0;

			&:hover {
				background: var(--accent-color-hover, rgba(59, 130, 246, 0.8));
				transform: translateY(-1px);
			}
		}
	}

	.credential-value {
		padding: 10px 15px 15px 15px;
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		font-family: "Courier New", monospace;
		font-size: 0.9em;
		word-break: break-all;
		line-height: 1.4;
		border-top: 1px solid var(--accent-color-light, rgba(59, 130, 246, 0.2));
		color: var(--text-color);

		&.monospace {
			letter-spacing: 1px;
			font-weight: 500;
		}
	}

	.setup-guide {
		margin-bottom: 20px;
		padding: 15px;
		background: var(--bg-color);
		border-radius: 8px;
		border-left: 4px solid var(--success-color, #22c55e);

		h4 {
			margin: 0 0 15px 0;
			font-size: 1.1em;
			color: var(--text-color);
		}

		.setup-steps {
			ol {
				margin: 0;
				padding-left: 20px;

				li {
					margin-bottom: 8px;
					line-height: 1.5;

					a {
						color: var(--accent-color);
						text-decoration: none;

						&:hover {
							text-decoration: underline;
						}
					}

					code {
						background: var(--bg-color);
						border: 1px solid var(--accent-color);
						padding: 2px 6px;
						border-radius: 4px;
						font-family: "Courier New", monospace;
						font-size: 0.9em;
					}
				}
			}
		}
	}

	.security-notice {
		padding: 15px;
		border-radius: 8px;
		background: rgba(245, 158, 11, 0.1);
		border: 1px solid rgba(245, 158, 11, 0.3);
		color: var(--warning-color, #f59e0b);

		h4 {
			margin: 0 0 10px 0;
			font-size: 1.05em;
		}

		p {
			margin: 0;
			line-height: 1.5;
			font-size: 0.95em;
		}
	}

	.credentials-modal .modal-footer {
		padding: 20px;
		border-top: 1px solid var(--accent-color);
		background: var(--bg-color);
	}

	.footer-actions {
		display: flex;
		gap: 10px;
		justify-content: flex-end;

		button {
			padding: 10px 20px;
			border-radius: 6px;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s ease;
			border: none;
			display: flex;
			align-items: center;
			gap: 8px;

			&.primary {
				background: var(--accent-color);
				color: white;

				&:hover {
					background: var(--accent-color-dark);
				}
			}

			&.warning {
				background: var(--warning-color, #f59e0b);
				color: white;

				&:hover {
					background: var(--warning-color-dark, #d97706);
				}
			}
		}
	}

	// Matrix Account Type Styles
	.matrix-account-type {
		background: var(--bg-secondary, rgba(255, 255, 255, 0.05));
		border-radius: 8px;
		padding: 15px;
		margin-bottom: 10px;

		.account-type-header {
			display: flex;
			align-items: center;
			gap: 12px;
			margin-bottom: 10px;

			.account-type-icon {
				font-size: 1.3em;
			}

			.account-type-info {
				display: flex;
				flex-direction: column;
				gap: 2px;

				.account-type-name {
					font-weight: 600;
					font-size: 1em;
					color: var(--text-color);
				}

				.account-type-desc {
					font-size: 0.85em;
					opacity: 0.8;
					color: var(--text-color);
				}
			}
		}

		.account-features {
			display: flex;
			flex-wrap: wrap;
			gap: 6px;

			.feature-tag {
				background: var(--accent-color);
				color: white;
				padding: 4px 8px;
				border-radius: 12px;
				font-size: 0.75em;
				font-weight: 500;
				white-space: nowrap;
			}
		}
	}

	.matrix-account-options {
		display: flex;
		flex-direction: column;
		gap: 20px;

		.account-option {
			border: 2px solid var(--border-color);
			border-radius: 12px;
			padding: 20px;
			background: var(--bg-color);
			transition: all 0.2s ease;

			&.recommended {
				border-color: var(--accent-color);
				background: var(--bg-secondary, rgba(59, 130, 246, 0.05));
			}

			&.disabled {
				border-color: var(--border-color);
				background: var(--bg-secondary, rgba(128, 128, 128, 0.05));
				opacity: 0.6;
				cursor: not-allowed;
			}

			&:hover:not(.disabled) {
				border-color: var(--accent-color);
				box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
			}

			.option-header {
				display: flex;
				align-items: center;
				gap: 12px;
				margin-bottom: 12px;

				.option-icon {
					font-size: 1.5em;
				}

				.option-info {
					display: flex;
					align-items: center;
					gap: 10px;
					flex: 1;

					.option-name {
						font-weight: 600;
						font-size: 1.1em;
						color: var(--text-color);
					}

					.option-badge {
						background: var(--accent-color);
						color: white;
						padding: 4px 8px;
						border-radius: 8px;
						font-size: 0.7em;
						font-weight: 600;
						text-transform: uppercase;
						letter-spacing: 0.5px;

						&.disabled {
							background: var(--text-color);
							opacity: 0.5;
						}
					}
				}
			}

			.option-description {
				margin: 0 0 15px 0;
				line-height: 1.5;
				color: var(--text-color);
				opacity: 0.9;
				font-size: 0.95em;
			}

			.option-features {
				display: flex;
				flex-wrap: wrap;
				gap: 8px;
				margin-bottom: 15px;

				.feature-tag {
					background: var(--bg-secondary, rgba(34, 197, 94, 0.1));
					color: var(--success-color, #22c55e);
					border: 1px solid var(--success-color, #22c55e);
					padding: 4px 8px;
					border-radius: 12px;
					font-size: 0.75em;
					font-weight: 500;
					white-space: nowrap;

					&.disabled {
						background: var(--bg-secondary, rgba(239, 68, 68, 0.1));
						color: var(--error-color, #ef4444);
						border: 1px solid var(--error-color, #ef4444);
						opacity: 0.7;
					}
				}
			}

			.option-button {
				padding: 10px 20px;
				border-radius: 8px;
				font-weight: 500;
				cursor: pointer;
				transition: all 0.2s ease;
				border: none;
				font-size: 0.9em;

				&.primary {
					background: var(--accent-color);
					color: white;

					&:hover:not(:disabled) {
						background: var(--accent-color-dark);
						transform: translateY(-1px);
					}
				}

				&.secondary {
					background: var(--bg-secondary);
					color: var(--text-color);
					border: 1px solid var(--border-color);

					&:hover:not(:disabled) {
						background: var(--accent-color);
						color: white;
						border-color: var(--accent-color);
					}
				}

				&:disabled {
					opacity: 0.6;
					cursor: not-allowed;
				}

				&.disabled {
					background: var(--bg-secondary, #6b7280) !important;
					color: var(--text-color) !important;
					border: 1px solid var(--border-color) !important;
					opacity: 0.5;
					cursor: not-allowed;
				}
			}

			.personal-account-options {
				display: flex;
				flex-direction: column;
				gap: 15px;
				margin-top: 15px;
				padding-top: 15px;
				border-top: 1px solid var(--border-color);

				.sub-option {
					padding: 15px;
					background: var(--bg-color);
					border-radius: 8px;
					border: 1px solid var(--border-color);

					h6 {
						margin: 0 0 5px 0;
						font-size: 0.9em;
						font-weight: 600;
						color: var(--text-color);
					}

					p {
						margin: 0 0 10px 0;
						font-size: 0.85em;
						opacity: 0.8;
						color: var(--text-color);

						&.note {
							font-size: 0.8em;
							opacity: 0.6;
							font-style: italic;
							margin-bottom: 5px;
						}
					}

					.link-form-inline {
						display: flex;
						flex-direction: column;
						gap: 8px;

						input {
							padding: 8px 12px;
							border: 1px solid var(--border-color);
							border-radius: 6px;
							background: var(--bg-color);
							color: var(--text-color);
							font-size: 0.9em;

							&:focus {
								outline: none;
								border-color: var(--accent-color);
								box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
							}

							&::placeholder {
								color: var(--text-color);
								opacity: 0.5;
							}
						}
					}
				}
			}
		}
	}

	@media (max-width: 768px) {
		.credentials-modal {
			width: 95%;
			max-height: 95vh;
		}

		.credentials-modal .modal-content {
			padding: 15px;
		}

		.credentials-modal .modal-header {
			padding: 15px;
		}

		.credential-header {
			flex-direction: column;
			align-items: stretch;
			gap: 10px;

			.credential-info {
				min-width: auto;
			}

			.copy-button {
				align-self: flex-end;
				width: fit-content;
			}
		}

		.footer-actions {
			flex-direction: column;

			button {
				width: 100%;
				justify-content: center;
			}
		}
	}
</style>
