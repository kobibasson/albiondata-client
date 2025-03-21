// Code generated by "stringer -type=OperationType"; DO NOT EDIT.

package client

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[opUnused-0]
	_ = x[opPing-1]
	_ = x[opJoin-2]
	_ = x[opVersionedOperation-3]
	_ = x[opCreateAccount-4]
	_ = x[opLogin-5]
	_ = x[opCreateGuestAccount-6]
	_ = x[opSendCrashLog-7]
	_ = x[opSendTraceRoute-8]
	_ = x[opSendVfxStats-9]
	_ = x[opSendGamePingInfo-10]
	_ = x[opCreateCharacter-11]
	_ = x[opDeleteCharacter-12]
	_ = x[opSelectCharacter-13]
	_ = x[opAcceptPopups-14]
	_ = x[opRedeemKeycode-15]
	_ = x[opGetGameServerByCluster-16]
	_ = x[opGetShopPurchaseUrl-17]
	_ = x[opGetReferralSeasonDetails-18]
	_ = x[opGetReferralLink-19]
	_ = x[opGetShopTilesForCategory-20]
	_ = x[opMove-21]
	_ = x[opAttackStart-22]
	_ = x[opCastStart-23]
	_ = x[opCastCancel-24]
	_ = x[opTerminateToggleSpell-25]
	_ = x[opChannelingCancel-26]
	_ = x[opAttackBuildingStart-27]
	_ = x[opInventoryDestroyItem-28]
	_ = x[opInventoryMoveItem-29]
	_ = x[opInventoryRecoverItem-30]
	_ = x[opInventoryRecoverAllItems-31]
	_ = x[opInventorySplitStack-32]
	_ = x[opInventorySplitStackInto-33]
	_ = x[opGetClusterData-34]
	_ = x[opChangeCluster-35]
	_ = x[opConsoleCommand-36]
	_ = x[opChatMessage-37]
	_ = x[opReportClientError-38]
	_ = x[opRegisterToObject-39]
	_ = x[opUnRegisterFromObject-40]
	_ = x[opCraftBuildingChangeSettings-41]
	_ = x[opCraftBuildingTakeMoney-42]
	_ = x[opRepairBuildingChangeSettings-43]
	_ = x[opRepairBuildingTakeMoney-44]
	_ = x[opActionBuildingChangeSettings-45]
	_ = x[opHarvestStart-46]
	_ = x[opHarvestCancel-47]
	_ = x[opTakeSilver-48]
	_ = x[opActionOnBuildingStart-49]
	_ = x[opActionOnBuildingCancel-50]
	_ = x[opInstallResourceStart-51]
	_ = x[opInstallResourceCancel-52]
	_ = x[opInstallSilver-53]
	_ = x[opBuildingFillNutrition-54]
	_ = x[opBuildingChangeRenovationState-55]
	_ = x[opBuildingBuySkin-56]
	_ = x[opBuildingClaim-57]
	_ = x[opBuildingGiveup-58]
	_ = x[opBuildingNutritionSilverStorageDeposit-59]
	_ = x[opBuildingNutritionSilverStorageWithdraw-60]
	_ = x[opBuildingNutritionSilverRewardSet-61]
	_ = x[opConstructionSiteCreate-62]
	_ = x[opPlaceableObjectPlace-63]
	_ = x[opPlaceableObjectPlaceCancel-64]
	_ = x[opPlaceableObjectPickup-65]
	_ = x[opFurnitureObjectUse-66]
	_ = x[opFarmableHarvest-67]
	_ = x[opFarmableFinishGrownItem-68]
	_ = x[opFarmableDestroy-69]
	_ = x[opFarmableGetProduct-70]
	_ = x[opFarmableFill-71]
	_ = x[opTearDownConstructionSite-72]
	_ = x[opAuctionCreateOffer-73]
	_ = x[opAuctionCreateRequest-74]
	_ = x[opAuctionGetOffers-75]
	_ = x[opAuctionGetRequests-76]
	_ = x[opAuctionBuyOffer-77]
	_ = x[opAuctionAbortAuction-78]
	_ = x[opAuctionModifyAuction-79]
	_ = x[opAuctionAbortOffer-80]
	_ = x[opAuctionAbortRequest-81]
	_ = x[opAuctionSellRequest-82]
	_ = x[opAuctionGetFinishedAuctions-83]
	_ = x[opAuctionGetFinishedAuctionsCount-84]
	_ = x[opAuctionFetchAuction-85]
	_ = x[opAuctionGetMyOpenOffers-86]
	_ = x[opAuctionGetMyOpenRequests-87]
	_ = x[opAuctionGetMyOpenAuctions-88]
	_ = x[opAuctionGetItemAverageStats-89]
	_ = x[opAuctionGetItemAverageValue-90]
	_ = x[opContainerOpen-91]
	_ = x[opContainerClose-92]
	_ = x[opContainerManageSubContainer-93]
	_ = x[opRespawn-94]
	_ = x[opSuicide-95]
	_ = x[opJoinGuild-96]
	_ = x[opLeaveGuild-97]
	_ = x[opCreateGuild-98]
	_ = x[opInviteToGuild-99]
	_ = x[opDeclineGuildInvitation-100]
	_ = x[opKickFromGuild-101]
	_ = x[opInstantJoinGuild-102]
	_ = x[opDuellingChallengePlayer-103]
	_ = x[opDuellingAcceptChallenge-104]
	_ = x[opDuellingDenyChallenge-105]
	_ = x[opChangeClusterTax-106]
	_ = x[opClaimTerritory-107]
	_ = x[opGiveUpTerritory-108]
	_ = x[opChangeTerritoryAccessRights-109]
	_ = x[opGetMonolithInfo-110]
	_ = x[opGetClaimInfo-111]
	_ = x[opGetAttackInfo-112]
	_ = x[opGetTerritorySeasonPoints-113]
	_ = x[opGetAttackSchedule-114]
	_ = x[opGetMatches-115]
	_ = x[opGetMatchDetails-116]
	_ = x[opJoinMatch-117]
	_ = x[opLeaveMatch-118]
	_ = x[opGetClusterInstanceInfoForStaticCluster-119]
	_ = x[opChangeChatSettings-120]
	_ = x[opLogoutStart-121]
	_ = x[opLogoutCancel-122]
	_ = x[opClaimOrbStart-123]
	_ = x[opClaimOrbCancel-124]
	_ = x[opMatchLootChestOpeningStart-125]
	_ = x[opMatchLootChestOpeningCancel-126]
	_ = x[opDepositToGuildAccount-127]
	_ = x[opWithdrawalFromAccount-128]
	_ = x[opChangeGuildPayUpkeepFlag-129]
	_ = x[opChangeGuildTax-130]
	_ = x[opGetMyTerritories-131]
	_ = x[opMorganaCommand-132]
	_ = x[opGetServerInfo-133]
	_ = x[opSubscribeToCluster-134]
	_ = x[opAnswerMercenaryInvitation-135]
	_ = x[opGetCharacterEquipment-136]
	_ = x[opGetCharacterSteamAchievements-137]
	_ = x[opGetCharacterStats-138]
	_ = x[opGetKillHistoryDetails-139]
	_ = x[opLearnMasteryLevel-140]
	_ = x[opReSpecAchievement-141]
	_ = x[opChangeAvatar-142]
	_ = x[opGetRankings-143]
	_ = x[opGetRank-144]
	_ = x[opGetGvgSeasonRankings-145]
	_ = x[opGetGvgSeasonRank-146]
	_ = x[opGetGvgSeasonHistoryRankings-147]
	_ = x[opGetGvgSeasonGuildMemberHistory-148]
	_ = x[opKickFromGvGMatch-149]
	_ = x[opGetCrystalLeagueDailySeasonPoints-150]
	_ = x[opGetChestLogs-151]
	_ = x[opGetAccessRightLogs-152]
	_ = x[opGetGuildAccountLogs-153]
	_ = x[opGetGuildAccountLogsLargeAmount-154]
	_ = x[opInviteToPlayerTrade-155]
	_ = x[opPlayerTradeCancel-156]
	_ = x[opPlayerTradeInvitationAccept-157]
	_ = x[opPlayerTradeAddItem-158]
	_ = x[opPlayerTradeRemoveItem-159]
	_ = x[opPlayerTradeAcceptTrade-160]
	_ = x[opPlayerTradeSetSilverOrGold-161]
	_ = x[opSendMiniMapPing-162]
	_ = x[opStuck-163]
	_ = x[opBuyRealEstate-164]
	_ = x[opClaimRealEstate-165]
	_ = x[opGiveUpRealEstate-166]
	_ = x[opChangeRealEstateOutline-167]
	_ = x[opGetMailInfos-168]
	_ = x[opGetMailCount-169]
	_ = x[opReadMail-170]
	_ = x[opSendNewMail-171]
	_ = x[opDeleteMail-172]
	_ = x[opMarkMailUnread-173]
	_ = x[opClaimAttachmentFromMail-174]
	_ = x[opApplyToGuild-175]
	_ = x[opAnswerGuildApplication-176]
	_ = x[opRequestGuildFinderFilteredList-177]
	_ = x[opUpdateGuildRecruitmentInfo-178]
	_ = x[opRequestGuildRecruitmentInfo-179]
	_ = x[opRequestGuildFinderNameSearch-180]
	_ = x[opRequestGuildFinderRecommendedList-181]
	_ = x[opRegisterChatPeer-182]
	_ = x[opSendChatMessage-183]
	_ = x[opSendModeratorMessage-184]
	_ = x[opJoinChatChannel-185]
	_ = x[opLeaveChatChannel-186]
	_ = x[opSendWhisperMessage-187]
	_ = x[opSay-188]
	_ = x[opPlayEmote-189]
	_ = x[opStopEmote-190]
	_ = x[opGetClusterMapInfo-191]
	_ = x[opAccessRightsChangeSettings-192]
	_ = x[opMount-193]
	_ = x[opMountCancel-194]
	_ = x[opBuyJourney-195]
	_ = x[opSetSaleStatusForEstate-196]
	_ = x[opResolveGuildOrPlayerName-197]
	_ = x[opGetRespawnInfos-198]
	_ = x[opMakeHome-199]
	_ = x[opLeaveHome-200]
	_ = x[opResurrectionReply-201]
	_ = x[opAllianceCreate-202]
	_ = x[opAllianceDisband-203]
	_ = x[opAllianceGetMemberInfos-204]
	_ = x[opAllianceInvite-205]
	_ = x[opAllianceAnswerInvitation-206]
	_ = x[opAllianceCancelInvitation-207]
	_ = x[opAllianceKickGuild-208]
	_ = x[opAllianceLeave-209]
	_ = x[opAllianceChangeGoldPaymentFlag-210]
	_ = x[opAllianceGetDetailInfo-211]
	_ = x[opGetIslandInfos-212]
	_ = x[opBuyMyIsland-213]
	_ = x[opBuyGuildIsland-214]
	_ = x[opUpgradeMyIsland-215]
	_ = x[opUpgradeGuildIsland-216]
	_ = x[opTerritoryFillNutrition-217]
	_ = x[opTeleportBack-218]
	_ = x[opPartyInvitePlayer-219]
	_ = x[opPartyRequestJoin-220]
	_ = x[opPartyAnswerInvitation-221]
	_ = x[opPartyAnswerJoinRequest-222]
	_ = x[opPartyLeave-223]
	_ = x[opPartyKickPlayer-224]
	_ = x[opPartyMakeLeader-225]
	_ = x[opPartyChangeLootSetting-226]
	_ = x[opPartyMarkObject-227]
	_ = x[opPartySetRole-228]
	_ = x[opSetGuildCodex-229]
	_ = x[opExitEnterStart-230]
	_ = x[opExitEnterCancel-231]
	_ = x[opQuestGiverRequest-232]
	_ = x[opGoldMarketGetBuyOffer-233]
	_ = x[opGoldMarketGetBuyOfferFromSilver-234]
	_ = x[opGoldMarketGetSellOffer-235]
	_ = x[opGoldMarketGetSellOfferFromSilver-236]
	_ = x[opGoldMarketBuyGold-237]
	_ = x[opGoldMarketSellGold-238]
	_ = x[opGoldMarketCreateSellOrder-239]
	_ = x[opGoldMarketCreateBuyOrder-240]
	_ = x[opGoldMarketGetInfos-241]
	_ = x[opGoldMarketCancelOrder-242]
	_ = x[opGoldMarketGetAverageInfo-243]
	_ = x[opTreasureChestUsingStart-244]
	_ = x[opTreasureChestUsingCancel-245]
	_ = x[opUseLootChest-246]
	_ = x[opUseShrine-247]
	_ = x[opUseHellgateShrine-248]
	_ = x[opGetSiegeBannerInfo-249]
	_ = x[opLaborerStartJob-250]
	_ = x[opLaborerTakeJobLoot-251]
	_ = x[opLaborerDismiss-252]
	_ = x[opLaborerMove-253]
	_ = x[opLaborerBuyItem-254]
	_ = x[opLaborerUpgrade-255]
	_ = x[opBuyPremium-256]
	_ = x[opRealEstateGetAuctionData-257]
	_ = x[opRealEstateBidOnAuction-258]
	_ = x[opFriendInvite-259]
	_ = x[opFriendAnswerInvitation-260]
	_ = x[opFriendCancelnvitation-261]
	_ = x[opFriendRemove-262]
	_ = x[opInventoryStack-263]
	_ = x[opInventoryReorder-264]
	_ = x[opInventoryDropAll-265]
	_ = x[opInventoryAddToStacks-266]
	_ = x[opEquipmentItemChangeSpell-267]
	_ = x[opExpeditionRegister-268]
	_ = x[opExpeditionRegisterCancel-269]
	_ = x[opJoinExpedition-270]
	_ = x[opDeclineExpeditionInvitation-271]
	_ = x[opVoteStart-272]
	_ = x[opVoteDoVote-273]
	_ = x[opRatingDoRate-274]
	_ = x[opEnteringExpeditionStart-275]
	_ = x[opEnteringExpeditionCancel-276]
	_ = x[opActivateExpeditionCheckPoint-277]
	_ = x[opArenaRegister-278]
	_ = x[opArenaAddInvite-279]
	_ = x[opArenaRegisterCancel-280]
	_ = x[opArenaLeave-281]
	_ = x[opJoinArenaMatch-282]
	_ = x[opDeclineArenaInvitation-283]
	_ = x[opEnteringArenaStart-284]
	_ = x[opEnteringArenaCancel-285]
	_ = x[opArenaCustomMatch-286]
	_ = x[opUpdateCharacterStatement-287]
	_ = x[opBoostFarmable-288]
	_ = x[opGetStrikeHistory-289]
	_ = x[opUseFunction-290]
	_ = x[opUsePortalEntrance-291]
	_ = x[opResetPortalBinding-292]
	_ = x[opQueryPortalBinding-293]
	_ = x[opClaimPaymentTransaction-294]
	_ = x[opChangeUseFlag-295]
	_ = x[opClientPerformanceStats-296]
	_ = x[opExtendedHardwareStats-297]
	_ = x[opClientLowMemoryWarning-298]
	_ = x[opTerritoryClaimStart-299]
	_ = x[opTerritoryClaimCancel-300]
	_ = x[opDeliverCarriableObjectStart-301]
	_ = x[opDeliverCarriableObjectCancel-302]
	_ = x[opTerritoryUpgradeWithPowerCrystal-303]
	_ = x[opRequestAppStoreProducts-304]
	_ = x[opVerifyProductPurchase-305]
	_ = x[opQueryGuildPlayerStats-306]
	_ = x[opQueryAllianceGuildStats-307]
	_ = x[opTrackAchievements-308]
	_ = x[opSetAchievementsAutoLearn-309]
	_ = x[opDepositItemToGuildCurrency-310]
	_ = x[opWithdrawalItemFromGuildCurrency-311]
	_ = x[opAuctionSellSpecificItemRequest-312]
	_ = x[opFishingStart-313]
	_ = x[opFishingCasting-314]
	_ = x[opFishingCast-315]
	_ = x[opFishingCatch-316]
	_ = x[opFishingPull-317]
	_ = x[opFishingGiveLine-318]
	_ = x[opFishingFinish-319]
	_ = x[opFishingCancel-320]
	_ = x[opCreateGuildAccessTag-321]
	_ = x[opDeleteGuildAccessTag-322]
	_ = x[opRenameGuildAccessTag-323]
	_ = x[opFlagGuildAccessTagGuildPermission-324]
	_ = x[opAssignGuildAccessTag-325]
	_ = x[opRemoveGuildAccessTagFromPlayer-326]
	_ = x[opModifyGuildAccessTagEditors-327]
	_ = x[opRequestPublicAccessTags-328]
	_ = x[opChangeAccessTagPublicFlag-329]
	_ = x[opUpdateGuildAccessTag-330]
	_ = x[opSteamStartMicrotransaction-331]
	_ = x[opSteamFinishMicrotransaction-332]
	_ = x[opSteamIdHasActiveAccount-333]
	_ = x[opCheckEmailAccountState-334]
	_ = x[opLinkAccountToSteamId-335]
	_ = x[opInAppConfirmPaymentGooglePlay-336]
	_ = x[opInAppConfirmPaymentAppleAppStore-337]
	_ = x[opInAppPurchaseRequest-338]
	_ = x[opInAppPurchaseFailed-339]
	_ = x[opCharacterSubscriptionInfo-340]
	_ = x[opAccountSubscriptionInfo-341]
	_ = x[opBuyGvgSeasonBooster-342]
	_ = x[opChangeFlaggingPrepare-343]
	_ = x[opOverCharge-344]
	_ = x[opOverChargeEnd-345]
	_ = x[opRequestTrusted-346]
	_ = x[opChangeGuildLogo-347]
	_ = x[opPartyFinderRegisterForUpdates-348]
	_ = x[opPartyFinderUnregisterForUpdates-349]
	_ = x[opPartyFinderEnlistNewPartySearch-350]
	_ = x[opPartyFinderDeletePartySearch-351]
	_ = x[opPartyFinderChangePartySearch-352]
	_ = x[opPartyFinderChangeRole-353]
	_ = x[opPartyFinderApplyForGroup-354]
	_ = x[opPartyFinderAcceptOrDeclineApplyForGroup-355]
	_ = x[opPartyFinderGetEquipmentSnapshot-356]
	_ = x[opPartyFinderRegisterApplicants-357]
	_ = x[opPartyFinderUnregisterApplicants-358]
	_ = x[opPartyFinderFulltextSearch-359]
	_ = x[opPartyFinderRequestEquipmentSnapshot-360]
	_ = x[opGetPersonalSeasonTrackerData-361]
	_ = x[opGetPersonalSeasonPastRewardData-362]
	_ = x[opUseConsumableFromInventory-363]
	_ = x[opClaimPersonalSeasonReward-364]
	_ = x[opEasyAntiCheatMessageToServer-365]
	_ = x[opXignCodeMessageToServer-366]
	_ = x[opBattlEyeMessageToServer-367]
	_ = x[opSetNextTutorialState-368]
	_ = x[opAddPlayerToMuteList-369]
	_ = x[opRemovePlayerFromMuteList-370]
	_ = x[opProductShopUserEvent-371]
	_ = x[opGetVanityUnlocks-372]
	_ = x[opBuyVanityUnlocks-373]
	_ = x[opGetMountSkins-374]
	_ = x[opSetMountSkin-375]
	_ = x[opSetWardrobe-376]
	_ = x[opChangeCustomization-377]
	_ = x[opChangePlayerIslandData-378]
	_ = x[opGetGuildChallengePoints-379]
	_ = x[opSmartQueueJoin-380]
	_ = x[opSmartQueueLeave-381]
	_ = x[opSmartQueueSelectSpawnCluster-382]
	_ = x[opUpgradeHideout-383]
	_ = x[opInitHideoutAttackStart-384]
	_ = x[opInitHideoutAttackCancel-385]
	_ = x[opHideoutFillNutrition-386]
	_ = x[opHideoutGetInfo-387]
	_ = x[opHideoutGetOwnerInfo-388]
	_ = x[opHideoutSetTribute-389]
	_ = x[opHideoutUpgradeWithPowerCrystal-390]
	_ = x[opHideoutDeclareHQ-391]
	_ = x[opHideoutUndeclareHQ-392]
	_ = x[opHideoutGetHQRequirements-393]
	_ = x[opHideoutBoost-394]
	_ = x[opHideoutBoostConstruction-395]
	_ = x[opOpenWorldAttackScheduleStart-396]
	_ = x[opOpenWorldAttackScheduleCancel-397]
	_ = x[opOpenWorldAttackConquerStart-398]
	_ = x[opOpenWorldAttackConquerCancel-399]
	_ = x[opGetOpenWorldAttackDetails-400]
	_ = x[opGetNextOpenWorldAttackScheduleTime-401]
	_ = x[opRecoverVaultFromHideout-402]
	_ = x[opGetGuildEnergyDrainInfo-403]
	_ = x[opChannelingUpdate-404]
	_ = x[opUseCorruptedShrine-405]
	_ = x[opRequestEstimatedMarketValue-406]
	_ = x[opLogFeedback-407]
	_ = x[opGetInfamyInfo-408]
	_ = x[opGetPartySmartClusterQueuePriority-409]
	_ = x[opSetPartySmartClusterQueuePriority-410]
	_ = x[opClientAntiAutoClickerInfo-411]
	_ = x[opClientBotPatternDetectionInfo-412]
	_ = x[opClientAntiGatherClickerInfo-413]
	_ = x[opLoadoutCreate-414]
	_ = x[opLoadoutRead-415]
	_ = x[opLoadoutReadHeaders-416]
	_ = x[opLoadoutUpdate-417]
	_ = x[opLoadoutDelete-418]
	_ = x[opLoadoutOrderUpdate-419]
	_ = x[opLoadoutEquip-420]
	_ = x[opBatchUseItemCancel-421]
	_ = x[opEnlistFactionWarfare-422]
	_ = x[opGetFactionWarfareWeeklyReport-423]
	_ = x[opClaimFactionWarfareWeeklyReport-424]
	_ = x[opGetFactionWarfareCampaignData-425]
	_ = x[opClaimFactionWarfareItemReward-426]
	_ = x[opSendMemoryConsumption-427]
	_ = x[opPickupCarriableObjectStart-428]
	_ = x[opPickupCarriableObjectCancel-429]
	_ = x[opSetSavingChestLogsFlag-430]
	_ = x[opGetSavingChestLogsFlag-431]
	_ = x[opRegisterGuestAccount-432]
	_ = x[opResendGuestAccountVerificationEmail-433]
	_ = x[opDoSimpleActionStart-434]
	_ = x[opDoSimpleActionCancel-435]
	_ = x[opGetGvgSeasonContributionByActivity-436]
	_ = x[opGetGvgSeasonContributionByCrystalLeague-437]
	_ = x[opGetGuildMightCategoryContribution-438]
	_ = x[opGetGuildMightCategoryOverview-439]
	_ = x[opGetPvpChallengeData-440]
	_ = x[opClaimPvpChallengeWeeklyReward-441]
	_ = x[opGetPersonalMightStats-442]
	_ = x[opAuctionGetLoadoutOffers-443]
	_ = x[opAuctionBuyLoadoutOffer-444]
	_ = x[opAccountDeletionRequest-445]
	_ = x[opAccountReactivationRequest-446]
	_ = x[opGetModerationEscalationDefiniton-447]
	_ = x[opEventBasedPopupAddSeen-448]
	_ = x[opGetItemKillHistory-449]
	_ = x[opGetVanityConsumables-450]
	_ = x[opEquipKillEmote-451]
	_ = x[opChangeKillEmotePlayOnKnockdownSetting-452]
	_ = x[opBuyVanityConsumableCharges-453]
	_ = x[opReclaimVanityItem-454]
	_ = x[opGetArenaRankings-455]
	_ = x[opGetCrystalLeagueStatistics-456]
	_ = x[opSendOptionsLog-457]
	_ = x[opSendControlsOptionsLog-458]
	_ = x[opMistsUseImmediateReturnExit-459]
	_ = x[opMistsUseStaticEntrance-460]
	_ = x[opMistsUseCityRoadsEntrance-461]
	_ = x[opChangeNewGuildMemberMail-462]
	_ = x[opGetNewGuildMemberMail-463]
	_ = x[opChangeGuildFactionAllegiance-464]
	_ = x[opGetGuildFactionAllegiance-465]
	_ = x[opGuildBannerChange-466]
	_ = x[opGuildGetOptionalStats-467]
	_ = x[opGuildSetOptionalStats-468]
	_ = x[opGetPlayerInfoForStalk-469]
	_ = x[opPayGoldForCharacterTypeChange-470]
	_ = x[opQuickSellAuctionQueryAction-471]
	_ = x[opQuickSellAuctionSellAction-472]
	_ = x[opFcmTokenToServer-473]
	_ = x[opApnsTokenToServer-474]
	_ = x[opDeathRecap-475]
	_ = x[opAuctionFetchFinishedAuctions-476]
	_ = x[opAbortAuctionFetchFinishedAuctions-477]
	_ = x[opRequestLegendaryEvenHistory-478]
	_ = x[opPartyAnswerStartHuntRequest-479]
	_ = x[opHuntAbort-480]
	_ = x[opUseFindTrackSpellFromItemPrepare-481]
	_ = x[opInteractWithTrackStart-482]
	_ = x[opInteractWithTrackCancel-483]
	_ = x[opTerritoryRaidStart-484]
	_ = x[opTerritoryRaidCancel-485]
	_ = x[opTerritoryClaimRaidedRawEnergyCrystalResult-486]
	_ = x[opGvGSeasonPlayerGuildParticipationDetails-487]
	_ = x[opDailyMightBonus-488]
	_ = x[opClaimDailyMightBonus-489]
	_ = x[opGetFortificationGroupInfo-490]
	_ = x[opUpgradeFortificationGroup-491]
	_ = x[opCancelUpgradeFortificationGroup-492]
	_ = x[opDowngradeFortificationGroup-493]
	_ = x[opGetClusterActivityChestEstimates-494]
	_ = x[opPartyReadyCheckBegin-495]
	_ = x[opPartyReadyCheckUpdate-496]
	_ = x[opClaimAlbionJournalReward-497]
	_ = x[opTrackAlbionJournalAchievements-498]
	_ = x[opRequestOutlandsTeleportationUsage-499]
	_ = x[opPickupFromPiledObjectStart-500]
	_ = x[opPickupFromPiledObjectCancel-501]
	_ = x[opAssetOverview-502]
	_ = x[opAssetOverviewTabs-503]
	_ = x[opAssetOverviewTabContent-504]
	_ = x[opAssetOverviewUnfreezeCache-505]
	_ = x[opAssetOverviewSearch-506]
	_ = x[opAssetOverviewSearchTabs-507]
	_ = x[opAssetOverviewSearchTabContent-508]
	_ = x[opAssetOverviewRecoverPlayerVault-509]
	_ = x[opImmortalizeKillTrophy-510]
}

const _OperationType_name = "opUnusedopPingopJoinopVersionedOperationopCreateAccountopLoginopCreateGuestAccountopSendCrashLogopSendTraceRouteopSendVfxStatsopSendGamePingInfoopCreateCharacteropDeleteCharacteropSelectCharacteropAcceptPopupsopRedeemKeycodeopGetGameServerByClusteropGetShopPurchaseUrlopGetReferralSeasonDetailsopGetReferralLinkopGetShopTilesForCategoryopMoveopAttackStartopCastStartopCastCancelopTerminateToggleSpellopChannelingCancelopAttackBuildingStartopInventoryDestroyItemopInventoryMoveItemopInventoryRecoverItemopInventoryRecoverAllItemsopInventorySplitStackopInventorySplitStackIntoopGetClusterDataopChangeClusteropConsoleCommandopChatMessageopReportClientErroropRegisterToObjectopUnRegisterFromObjectopCraftBuildingChangeSettingsopCraftBuildingTakeMoneyopRepairBuildingChangeSettingsopRepairBuildingTakeMoneyopActionBuildingChangeSettingsopHarvestStartopHarvestCancelopTakeSilveropActionOnBuildingStartopActionOnBuildingCancelopInstallResourceStartopInstallResourceCancelopInstallSilveropBuildingFillNutritionopBuildingChangeRenovationStateopBuildingBuySkinopBuildingClaimopBuildingGiveupopBuildingNutritionSilverStorageDepositopBuildingNutritionSilverStorageWithdrawopBuildingNutritionSilverRewardSetopConstructionSiteCreateopPlaceableObjectPlaceopPlaceableObjectPlaceCancelopPlaceableObjectPickupopFurnitureObjectUseopFarmableHarvestopFarmableFinishGrownItemopFarmableDestroyopFarmableGetProductopFarmableFillopTearDownConstructionSiteopAuctionCreateOfferopAuctionCreateRequestopAuctionGetOffersopAuctionGetRequestsopAuctionBuyOfferopAuctionAbortAuctionopAuctionModifyAuctionopAuctionAbortOfferopAuctionAbortRequestopAuctionSellRequestopAuctionGetFinishedAuctionsopAuctionGetFinishedAuctionsCountopAuctionFetchAuctionopAuctionGetMyOpenOffersopAuctionGetMyOpenRequestsopAuctionGetMyOpenAuctionsopAuctionGetItemAverageStatsopAuctionGetItemAverageValueopContainerOpenopContainerCloseopContainerManageSubContaineropRespawnopSuicideopJoinGuildopLeaveGuildopCreateGuildopInviteToGuildopDeclineGuildInvitationopKickFromGuildopInstantJoinGuildopDuellingChallengePlayeropDuellingAcceptChallengeopDuellingDenyChallengeopChangeClusterTaxopClaimTerritoryopGiveUpTerritoryopChangeTerritoryAccessRightsopGetMonolithInfoopGetClaimInfoopGetAttackInfoopGetTerritorySeasonPointsopGetAttackScheduleopGetMatchesopGetMatchDetailsopJoinMatchopLeaveMatchopGetClusterInstanceInfoForStaticClusteropChangeChatSettingsopLogoutStartopLogoutCancelopClaimOrbStartopClaimOrbCancelopMatchLootChestOpeningStartopMatchLootChestOpeningCancelopDepositToGuildAccountopWithdrawalFromAccountopChangeGuildPayUpkeepFlagopChangeGuildTaxopGetMyTerritoriesopMorganaCommandopGetServerInfoopSubscribeToClusteropAnswerMercenaryInvitationopGetCharacterEquipmentopGetCharacterSteamAchievementsopGetCharacterStatsopGetKillHistoryDetailsopLearnMasteryLevelopReSpecAchievementopChangeAvataropGetRankingsopGetRankopGetGvgSeasonRankingsopGetGvgSeasonRankopGetGvgSeasonHistoryRankingsopGetGvgSeasonGuildMemberHistoryopKickFromGvGMatchopGetCrystalLeagueDailySeasonPointsopGetChestLogsopGetAccessRightLogsopGetGuildAccountLogsopGetGuildAccountLogsLargeAmountopInviteToPlayerTradeopPlayerTradeCancelopPlayerTradeInvitationAcceptopPlayerTradeAddItemopPlayerTradeRemoveItemopPlayerTradeAcceptTradeopPlayerTradeSetSilverOrGoldopSendMiniMapPingopStuckopBuyRealEstateopClaimRealEstateopGiveUpRealEstateopChangeRealEstateOutlineopGetMailInfosopGetMailCountopReadMailopSendNewMailopDeleteMailopMarkMailUnreadopClaimAttachmentFromMailopApplyToGuildopAnswerGuildApplicationopRequestGuildFinderFilteredListopUpdateGuildRecruitmentInfoopRequestGuildRecruitmentInfoopRequestGuildFinderNameSearchopRequestGuildFinderRecommendedListopRegisterChatPeeropSendChatMessageopSendModeratorMessageopJoinChatChannelopLeaveChatChannelopSendWhisperMessageopSayopPlayEmoteopStopEmoteopGetClusterMapInfoopAccessRightsChangeSettingsopMountopMountCancelopBuyJourneyopSetSaleStatusForEstateopResolveGuildOrPlayerNameopGetRespawnInfosopMakeHomeopLeaveHomeopResurrectionReplyopAllianceCreateopAllianceDisbandopAllianceGetMemberInfosopAllianceInviteopAllianceAnswerInvitationopAllianceCancelInvitationopAllianceKickGuildopAllianceLeaveopAllianceChangeGoldPaymentFlagopAllianceGetDetailInfoopGetIslandInfosopBuyMyIslandopBuyGuildIslandopUpgradeMyIslandopUpgradeGuildIslandopTerritoryFillNutritionopTeleportBackopPartyInvitePlayeropPartyRequestJoinopPartyAnswerInvitationopPartyAnswerJoinRequestopPartyLeaveopPartyKickPlayeropPartyMakeLeaderopPartyChangeLootSettingopPartyMarkObjectopPartySetRoleopSetGuildCodexopExitEnterStartopExitEnterCancelopQuestGiverRequestopGoldMarketGetBuyOfferopGoldMarketGetBuyOfferFromSilveropGoldMarketGetSellOfferopGoldMarketGetSellOfferFromSilveropGoldMarketBuyGoldopGoldMarketSellGoldopGoldMarketCreateSellOrderopGoldMarketCreateBuyOrderopGoldMarketGetInfosopGoldMarketCancelOrderopGoldMarketGetAverageInfoopTreasureChestUsingStartopTreasureChestUsingCancelopUseLootChestopUseShrineopUseHellgateShrineopGetSiegeBannerInfoopLaborerStartJobopLaborerTakeJobLootopLaborerDismissopLaborerMoveopLaborerBuyItemopLaborerUpgradeopBuyPremiumopRealEstateGetAuctionDataopRealEstateBidOnAuctionopFriendInviteopFriendAnswerInvitationopFriendCancelnvitationopFriendRemoveopInventoryStackopInventoryReorderopInventoryDropAllopInventoryAddToStacksopEquipmentItemChangeSpellopExpeditionRegisteropExpeditionRegisterCancelopJoinExpeditionopDeclineExpeditionInvitationopVoteStartopVoteDoVoteopRatingDoRateopEnteringExpeditionStartopEnteringExpeditionCancelopActivateExpeditionCheckPointopArenaRegisteropArenaAddInviteopArenaRegisterCancelopArenaLeaveopJoinArenaMatchopDeclineArenaInvitationopEnteringArenaStartopEnteringArenaCancelopArenaCustomMatchopUpdateCharacterStatementopBoostFarmableopGetStrikeHistoryopUseFunctionopUsePortalEntranceopResetPortalBindingopQueryPortalBindingopClaimPaymentTransactionopChangeUseFlagopClientPerformanceStatsopExtendedHardwareStatsopClientLowMemoryWarningopTerritoryClaimStartopTerritoryClaimCancelopDeliverCarriableObjectStartopDeliverCarriableObjectCancelopTerritoryUpgradeWithPowerCrystalopRequestAppStoreProductsopVerifyProductPurchaseopQueryGuildPlayerStatsopQueryAllianceGuildStatsopTrackAchievementsopSetAchievementsAutoLearnopDepositItemToGuildCurrencyopWithdrawalItemFromGuildCurrencyopAuctionSellSpecificItemRequestopFishingStartopFishingCastingopFishingCastopFishingCatchopFishingPullopFishingGiveLineopFishingFinishopFishingCancelopCreateGuildAccessTagopDeleteGuildAccessTagopRenameGuildAccessTagopFlagGuildAccessTagGuildPermissionopAssignGuildAccessTagopRemoveGuildAccessTagFromPlayeropModifyGuildAccessTagEditorsopRequestPublicAccessTagsopChangeAccessTagPublicFlagopUpdateGuildAccessTagopSteamStartMicrotransactionopSteamFinishMicrotransactionopSteamIdHasActiveAccountopCheckEmailAccountStateopLinkAccountToSteamIdopInAppConfirmPaymentGooglePlayopInAppConfirmPaymentAppleAppStoreopInAppPurchaseRequestopInAppPurchaseFailedopCharacterSubscriptionInfoopAccountSubscriptionInfoopBuyGvgSeasonBoosteropChangeFlaggingPrepareopOverChargeopOverChargeEndopRequestTrustedopChangeGuildLogoopPartyFinderRegisterForUpdatesopPartyFinderUnregisterForUpdatesopPartyFinderEnlistNewPartySearchopPartyFinderDeletePartySearchopPartyFinderChangePartySearchopPartyFinderChangeRoleopPartyFinderApplyForGroupopPartyFinderAcceptOrDeclineApplyForGroupopPartyFinderGetEquipmentSnapshotopPartyFinderRegisterApplicantsopPartyFinderUnregisterApplicantsopPartyFinderFulltextSearchopPartyFinderRequestEquipmentSnapshotopGetPersonalSeasonTrackerDataopGetPersonalSeasonPastRewardDataopUseConsumableFromInventoryopClaimPersonalSeasonRewardopEasyAntiCheatMessageToServeropXignCodeMessageToServeropBattlEyeMessageToServeropSetNextTutorialStateopAddPlayerToMuteListopRemovePlayerFromMuteListopProductShopUserEventopGetVanityUnlocksopBuyVanityUnlocksopGetMountSkinsopSetMountSkinopSetWardrobeopChangeCustomizationopChangePlayerIslandDataopGetGuildChallengePointsopSmartQueueJoinopSmartQueueLeaveopSmartQueueSelectSpawnClusteropUpgradeHideoutopInitHideoutAttackStartopInitHideoutAttackCancelopHideoutFillNutritionopHideoutGetInfoopHideoutGetOwnerInfoopHideoutSetTributeopHideoutUpgradeWithPowerCrystalopHideoutDeclareHQopHideoutUndeclareHQopHideoutGetHQRequirementsopHideoutBoostopHideoutBoostConstructionopOpenWorldAttackScheduleStartopOpenWorldAttackScheduleCancelopOpenWorldAttackConquerStartopOpenWorldAttackConquerCancelopGetOpenWorldAttackDetailsopGetNextOpenWorldAttackScheduleTimeopRecoverVaultFromHideoutopGetGuildEnergyDrainInfoopChannelingUpdateopUseCorruptedShrineopRequestEstimatedMarketValueopLogFeedbackopGetInfamyInfoopGetPartySmartClusterQueuePriorityopSetPartySmartClusterQueuePriorityopClientAntiAutoClickerInfoopClientBotPatternDetectionInfoopClientAntiGatherClickerInfoopLoadoutCreateopLoadoutReadopLoadoutReadHeadersopLoadoutUpdateopLoadoutDeleteopLoadoutOrderUpdateopLoadoutEquipopBatchUseItemCancelopEnlistFactionWarfareopGetFactionWarfareWeeklyReportopClaimFactionWarfareWeeklyReportopGetFactionWarfareCampaignDataopClaimFactionWarfareItemRewardopSendMemoryConsumptionopPickupCarriableObjectStartopPickupCarriableObjectCancelopSetSavingChestLogsFlagopGetSavingChestLogsFlagopRegisterGuestAccountopResendGuestAccountVerificationEmailopDoSimpleActionStartopDoSimpleActionCancelopGetGvgSeasonContributionByActivityopGetGvgSeasonContributionByCrystalLeagueopGetGuildMightCategoryContributionopGetGuildMightCategoryOverviewopGetPvpChallengeDataopClaimPvpChallengeWeeklyRewardopGetPersonalMightStatsopAuctionGetLoadoutOffersopAuctionBuyLoadoutOfferopAccountDeletionRequestopAccountReactivationRequestopGetModerationEscalationDefinitonopEventBasedPopupAddSeenopGetItemKillHistoryopGetVanityConsumablesopEquipKillEmoteopChangeKillEmotePlayOnKnockdownSettingopBuyVanityConsumableChargesopReclaimVanityItemopGetArenaRankingsopGetCrystalLeagueStatisticsopSendOptionsLogopSendControlsOptionsLogopMistsUseImmediateReturnExitopMistsUseStaticEntranceopMistsUseCityRoadsEntranceopChangeNewGuildMemberMailopGetNewGuildMemberMailopChangeGuildFactionAllegianceopGetGuildFactionAllegianceopGuildBannerChangeopGuildGetOptionalStatsopGuildSetOptionalStatsopGetPlayerInfoForStalkopPayGoldForCharacterTypeChangeopQuickSellAuctionQueryActionopQuickSellAuctionSellActionopFcmTokenToServeropApnsTokenToServeropDeathRecapopAuctionFetchFinishedAuctionsopAbortAuctionFetchFinishedAuctionsopRequestLegendaryEvenHistoryopPartyAnswerStartHuntRequestopHuntAbortopUseFindTrackSpellFromItemPrepareopInteractWithTrackStartopInteractWithTrackCancelopTerritoryRaidStartopTerritoryRaidCancelopTerritoryClaimRaidedRawEnergyCrystalResultopGvGSeasonPlayerGuildParticipationDetailsopDailyMightBonusopClaimDailyMightBonusopGetFortificationGroupInfoopUpgradeFortificationGroupopCancelUpgradeFortificationGroupopDowngradeFortificationGroupopGetClusterActivityChestEstimatesopPartyReadyCheckBeginopPartyReadyCheckUpdateopClaimAlbionJournalRewardopTrackAlbionJournalAchievementsopRequestOutlandsTeleportationUsageopPickupFromPiledObjectStartopPickupFromPiledObjectCancelopAssetOverviewopAssetOverviewTabsopAssetOverviewTabContentopAssetOverviewUnfreezeCacheopAssetOverviewSearchopAssetOverviewSearchTabsopAssetOverviewSearchTabContentopAssetOverviewRecoverPlayerVaultopImmortalizeKillTrophy"

var _OperationType_index = [...]uint16{0, 8, 14, 20, 40, 55, 62, 82, 96, 112, 126, 144, 161, 178, 195, 209, 224, 248, 268, 294, 311, 336, 342, 355, 366, 378, 400, 418, 439, 461, 480, 502, 528, 549, 574, 590, 605, 621, 634, 653, 671, 693, 722, 746, 776, 801, 831, 845, 860, 872, 895, 919, 941, 964, 979, 1002, 1033, 1050, 1065, 1081, 1120, 1160, 1194, 1218, 1240, 1268, 1291, 1311, 1328, 1353, 1370, 1390, 1404, 1430, 1450, 1472, 1490, 1510, 1527, 1548, 1570, 1589, 1610, 1630, 1658, 1691, 1712, 1736, 1762, 1788, 1816, 1844, 1859, 1875, 1904, 1913, 1922, 1933, 1945, 1958, 1973, 1997, 2012, 2030, 2055, 2080, 2103, 2121, 2137, 2154, 2183, 2200, 2214, 2229, 2255, 2274, 2286, 2303, 2314, 2326, 2366, 2386, 2399, 2413, 2428, 2444, 2472, 2501, 2524, 2547, 2573, 2589, 2607, 2623, 2638, 2658, 2685, 2708, 2739, 2758, 2781, 2800, 2819, 2833, 2846, 2855, 2877, 2895, 2924, 2956, 2974, 3009, 3023, 3043, 3064, 3096, 3117, 3136, 3165, 3185, 3208, 3232, 3260, 3277, 3284, 3299, 3316, 3334, 3359, 3373, 3387, 3397, 3410, 3422, 3438, 3463, 3477, 3501, 3533, 3561, 3590, 3620, 3655, 3673, 3690, 3712, 3729, 3747, 3767, 3772, 3783, 3794, 3813, 3841, 3848, 3861, 3873, 3897, 3923, 3940, 3950, 3961, 3980, 3996, 4013, 4037, 4053, 4079, 4105, 4124, 4139, 4170, 4193, 4209, 4222, 4238, 4255, 4275, 4299, 4313, 4332, 4350, 4373, 4397, 4409, 4426, 4443, 4467, 4484, 4498, 4513, 4529, 4546, 4565, 4588, 4621, 4645, 4679, 4698, 4718, 4745, 4771, 4791, 4814, 4840, 4865, 4891, 4905, 4916, 4935, 4955, 4972, 4992, 5008, 5021, 5037, 5053, 5065, 5091, 5115, 5129, 5153, 5176, 5190, 5206, 5224, 5242, 5264, 5290, 5310, 5336, 5352, 5381, 5392, 5404, 5418, 5443, 5469, 5499, 5514, 5530, 5551, 5563, 5579, 5603, 5623, 5644, 5662, 5688, 5703, 5721, 5734, 5753, 5773, 5793, 5818, 5833, 5857, 5880, 5904, 5925, 5947, 5976, 6006, 6040, 6065, 6088, 6111, 6136, 6155, 6181, 6209, 6242, 6274, 6288, 6304, 6317, 6331, 6344, 6361, 6376, 6391, 6413, 6435, 6457, 6492, 6514, 6546, 6575, 6600, 6627, 6649, 6677, 6706, 6731, 6755, 6777, 6808, 6842, 6864, 6885, 6912, 6937, 6958, 6981, 6993, 7008, 7024, 7041, 7072, 7105, 7138, 7168, 7198, 7221, 7247, 7288, 7321, 7352, 7385, 7412, 7449, 7479, 7512, 7540, 7567, 7597, 7622, 7647, 7669, 7690, 7716, 7738, 7756, 7774, 7789, 7803, 7816, 7837, 7861, 7886, 7902, 7919, 7949, 7965, 7989, 8014, 8036, 8052, 8073, 8092, 8124, 8142, 8162, 8188, 8202, 8228, 8258, 8289, 8318, 8348, 8375, 8411, 8436, 8461, 8479, 8499, 8528, 8541, 8556, 8591, 8626, 8653, 8684, 8713, 8728, 8741, 8761, 8776, 8791, 8811, 8825, 8845, 8867, 8898, 8931, 8962, 8993, 9016, 9044, 9073, 9097, 9121, 9143, 9180, 9201, 9223, 9259, 9300, 9335, 9366, 9387, 9418, 9441, 9466, 9490, 9514, 9542, 9576, 9600, 9620, 9642, 9658, 9697, 9725, 9744, 9762, 9790, 9806, 9830, 9859, 9883, 9910, 9936, 9959, 9989, 10016, 10035, 10058, 10081, 10104, 10135, 10164, 10192, 10210, 10229, 10241, 10271, 10306, 10335, 10364, 10375, 10409, 10433, 10458, 10478, 10499, 10543, 10585, 10602, 10624, 10651, 10678, 10711, 10740, 10774, 10796, 10819, 10845, 10877, 10912, 10940, 10969, 10984, 11003, 11028, 11056, 11077, 11102, 11133, 11166, 11189}

func (i OperationType) String() string {
	if i >= OperationType(len(_OperationType_index)-1) {
		return "OperationType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _OperationType_name[_OperationType_index[i]:_OperationType_index[i+1]]
}
