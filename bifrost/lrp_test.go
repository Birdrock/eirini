package bifrost_test

import (
	"context"
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/eirini/bifrost"
	"code.cloudfoundry.org/eirini/bifrost/bifrostfakes"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bifrost LRP", func() {

	var (
		err          error
		lrpBifrost   *bifrost.LRP
		request      cf.DesireLRPRequest
		lrpConverter *bifrostfakes.FakeLRPConverter
		lrpDesirer   *bifrostfakes.FakeLRPDesirer
	)

	BeforeEach(func() {
		lrpConverter = new(bifrostfakes.FakeLRPConverter)
		lrpDesirer = new(bifrostfakes.FakeLRPDesirer)
	})

	JustBeforeEach(func() {
		lrpBifrost = &bifrost.LRP{
			Converter: lrpConverter,
			Desirer:   lrpDesirer,
		}
	})

	Describe("Transfer LRP", func() {

		Context("When lrp is transferred successfully", func() {
			var lrp opi.LRP
			JustBeforeEach(func() {
				err = lrpBifrost.Transfer(context.Background(), request)
			})

			BeforeEach(func() {
				lrp = opi.LRP{
					Image: "docker.png",
				}
				lrpConverter.ConvertLRPReturns(lrp, nil)
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should use Converter", func() {
				Expect(lrpConverter.ConvertLRPCallCount()).To(Equal(1))
				Expect(lrpConverter.ConvertLRPArgsForCall(0)).To(Equal(request))
			})

			It("should use Desirer with the converted LRP", func() {
				Expect(lrpDesirer.DesireCallCount()).To(Equal(1))
				desired := lrpDesirer.DesireArgsForCall(0)
				Expect(desired).To(Equal(&lrp))
			})
		})

		Context("When lrp transfer fails", func() {
			Context("when Converter fails", func() {
				BeforeEach(func() {
					request = cf.DesireLRPRequest{GUID: "my-guid"}
					lrpConverter.ConvertLRPReturns(opi.LRP{}, errors.New("failed-to-convert"))
				})

				It("should return an error", func() {
					Expect(lrpBifrost.Transfer(context.Background(), request)).To(MatchError(ContainSubstring("failed to convert request")))
				})

				It("should use Converter", func() {
					Expect(lrpBifrost.Transfer(context.Background(), request)).ToNot(Succeed())
					Expect(lrpConverter.ConvertLRPCallCount()).To(Equal(1))
					Expect(lrpConverter.ConvertLRPArgsForCall(0)).To(Equal(request))
				})

				It("should not use Desirer", func() {
					Expect(lrpBifrost.Transfer(context.Background(), request)).ToNot(Succeed())
					Expect(lrpDesirer.DesireCallCount()).To(Equal(0))
				})

			})

			Context("When Desirer fails", func() {
				BeforeEach(func() {
					lrpDesirer.DesireReturns(errors.New("failed-to-desire-main-error"))
				})

				It("shoud return an error", func() {
					Expect(lrpBifrost.Transfer(context.Background(), request)).To(MatchError(ContainSubstring("failed to desire")))
				})
			})
		})

	})

	Describe("List LRP", func() {
		createLRP := func(processGUID, version, lastUpdated string) *opi.LRP {
			return &opi.LRP{
				LRPIdentifier: opi.LRPIdentifier{GUID: processGUID, Version: version},
				LastUpdated:   lastUpdated,
			}
		}

		Context("When no running LRPs exist", func() {
			It("should return an empty list of DesiredLRPSchedulingInfo", func() {
				Expect(lrpBifrost.List(context.Background())).To(HaveLen(0))
			})
		})

		Context("When listing running LRPs", func() {
			BeforeEach(func() {
				lrps := []*opi.LRP{
					createLRP("abcd", "123", "3464634.2"),
					createLRP("efgh", "234", "235.26535"),
					createLRP("ijkl", "123", "2342342.2"),
				}
				lrpDesirer.ListReturns(lrps, nil)
			})

			It("should succeed", func() {
				_, listErr := lrpBifrost.List(context.Background())
				Expect(listErr).ToNot(HaveOccurred())
			})

			It("should translate []LRPs to []DesiredLRPSchedulingInfo", func() {
				desiredLRPSchedulingInfos, _ := lrpBifrost.List(context.Background())
				Expect(desiredLRPSchedulingInfos).To(HaveLen(3))
				Expect(desiredLRPSchedulingInfos[0].ProcessGuid).To(Equal("abcd-123"))
				Expect(desiredLRPSchedulingInfos[1].ProcessGuid).To(Equal("efgh-234"))
				Expect(desiredLRPSchedulingInfos[2].ProcessGuid).To(Equal("ijkl-123"))

				Expect(desiredLRPSchedulingInfos[0].Annotation).To(Equal("3464634.2"))
				Expect(desiredLRPSchedulingInfos[1].Annotation).To(Equal("235.26535"))
				Expect(desiredLRPSchedulingInfos[2].Annotation).To(Equal("2342342.2"))
			})
		})

		Context("When an error occurs", func() {
			BeforeEach(func() {
				lrpDesirer.ListReturns(nil, errors.New("arrgh"))
			})

			It("should return a meaningful errormessage", func() {
				_, listErr := lrpBifrost.List(context.Background())
				Expect(listErr).To(MatchError(ContainSubstring("failed to list desired LRPs")))
			})
		})
	})

	Describe("Update an app", func() {

		var (
			updateRequest cf.UpdateDesiredLRPRequest
		)

		BeforeEach(func() {
			updateRequest = cf.UpdateDesiredLRPRequest{
				GUID:    "guid_1234",
				Version: "version_1234",
			}
		})

		JustBeforeEach(func() {
			err = lrpBifrost.Update(context.Background(), updateRequest)
		})

		Context("when the app exists", func() {

			BeforeEach(func() {
				lrp := opi.LRP{
					TargetInstances: 2,
					LastUpdated:     "whenever",
					AppURIs:         `[{"hostname":"my.route","port":8080},{"hostname":"your.route","port":5555}]`,
				}
				lrpDesirer.GetReturns(&lrp, nil)
			})

			Context("with instance count modified", func() {
				BeforeEach(func() {
					updateRequest.Update = &models.DesiredLRPUpdate{}
					updateRequest.Update.SetInstances(int32(5))
					updateRequest.Update.SetAnnotation("21421321.3")
					lrpDesirer.UpdateReturns(nil)
				})

				It("should get the existing LRP", func() {
					Expect(lrpDesirer.GetCallCount()).To(Equal(1))
					identifier := lrpDesirer.GetArgsForCall(0)
					Expect(identifier.GUID).To(Equal("guid_1234"))
					Expect(identifier.Version).To(Equal("version_1234"))
				})

				It("should submit the updated LRP", func() {
					Expect(lrpDesirer.UpdateCallCount()).To(Equal(1))
					lrp := lrpDesirer.UpdateArgsForCall(0)
					Expect(lrp.TargetInstances).To(Equal(int(5)))
					Expect(lrp.LastUpdated).To(Equal("21421321.3"))
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when the update fails", func() {
					BeforeEach(func() {
						lrpDesirer.UpdateReturns(errors.New("your app is bad"))
					})

					It("should propagate the error", func() {
						Expect(err).To(MatchError(ContainSubstring("failed to update")))
					})
				})
			})

			Context("When the routes are updated", func() {
				BeforeEach(func() {
					updatedRoutes := []map[string]interface{}{
						{
							"hostname": "my.route",
							"port":     8080,
						},
						{
							"hostname": "my.other.route",
							"port":     7777,
						},
					}

					routesJSON, marshalErr := json.Marshal(updatedRoutes)
					Expect(marshalErr).ToNot(HaveOccurred())

					rawJSON := json.RawMessage(routesJSON)

					updatedInstances := int32(5)
					updatedTimestamp := "23456.7"
					updateRequest.Update = &models.DesiredLRPUpdate{
						Routes: &models.Routes{
							"cf-router": &rawJSON,
						},
					}
					updateRequest.Update.SetInstances(updatedInstances)
					updateRequest.Update.SetAnnotation(updatedTimestamp)

					lrpDesirer.UpdateReturns(nil)
				})

				It("should get the existing LRP", func() {
					Expect(lrpDesirer.GetCallCount()).To(Equal(1))
					identifier := lrpDesirer.GetArgsForCall(0)
					Expect(identifier.GUID).To(Equal("guid_1234"))
					Expect(identifier.Version).To(Equal("version_1234"))
				})

				It("should have the updated routes", func() {
					Expect(lrpDesirer.UpdateCallCount()).To(Equal(1))
					lrp := lrpDesirer.UpdateArgsForCall(0)
					Expect(lrp.AppURIs).To(Equal(`[{"hostname":"my.route","port":8080},{"hostname":"my.other.route","port":7777}]`))
				})

				Context("When there are no routes provided", func() {
					BeforeEach(func() {
						updatedRoutes := []map[string]interface{}{}

						routesJSON, marshalErr := json.Marshal(updatedRoutes)
						Expect(marshalErr).ToNot(HaveOccurred())

						rawJSON := json.RawMessage(routesJSON)
						updateRequest.Update.Routes = &models.Routes{
							"cf-router": &rawJSON,
						}
					})

					It("should update it to an empty array", func() {
						Expect(lrpDesirer.UpdateCallCount()).To(Equal(1))
						lrp := lrpDesirer.UpdateArgsForCall(0)
						Expect(lrp.AppURIs).To(Equal(`[]`))
					})
				})
			})
		})

		Context("when the app does not exist", func() {

			BeforeEach(func() {
				lrpDesirer.GetReturns(nil, errors.New("app does not exist"))
			})

			It("should try to get the LRP", func() {
				Expect(lrpDesirer.GetCallCount()).To(Equal(1))
				identifier := lrpDesirer.GetArgsForCall(0)
				Expect(identifier.GUID).To(Equal("guid_1234"))
				Expect(identifier.Version).To(Equal("version_1234"))

			})

			It("should not submit anything to be updated", func() {
				Expect(lrpDesirer.UpdateCallCount()).To(Equal(0))
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to get app")))
			})
		})
	})

	Describe("Get an App", func() {
		var (
			lrp        *opi.LRP
			identifier opi.LRPIdentifier
		)

		JustBeforeEach(func() {
			identifier = opi.LRPIdentifier{
				GUID:    "guid_1234",
				Version: "version_1234",
			}

		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				lrp = &opi.LRP{
					TargetInstances: 5,
				}

				lrpDesirer.GetReturns(lrp, nil)
			})

			It("should use the desirer to get the lrp", func() {
				_, err = lrpBifrost.GetApp(context.Background(), identifier)
				Expect(err).NotTo(HaveOccurred())
				Expect(lrpDesirer.GetCallCount()).To(Equal(1))
				Expect(lrpDesirer.GetArgsForCall(0)).To(Equal(identifier))
			})

			It("should return a DesiredLRP", func() {
				desiredLRP, _ := lrpBifrost.GetApp(context.Background(), identifier)
				Expect(desiredLRP).ToNot(BeNil())
				Expect(desiredLRP.ProcessGuid).To(Equal("guid_1234-version_1234"))
				Expect(desiredLRP.Instances).To(Equal(int32(5)))
			})

		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				lrpDesirer.GetReturns(nil, errors.New("Failed to get LRP"))
			})

			It("should return an error", func() {
				_, getAppErr := lrpBifrost.GetApp(context.Background(), identifier)
				Expect(getAppErr).To(MatchError(ContainSubstring("failed to get app")))
			})
		})
	})

	Describe("Stop an app", func() {

		JustBeforeEach(func() {
			err = lrpBifrost.Stop(context.Background(), opi.LRPIdentifier{GUID: "guid_1234", Version: "version_1234"})
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call the desirer with the expected guid", func() {
			identifier := lrpDesirer.StopArgsForCall(0)
			Expect(identifier.GUID).To(Equal("guid_1234"))
			Expect(identifier.Version).To(Equal("version_1234"))
		})

		Context("when desirer's stop fails", func() {

			BeforeEach(func() {
				lrpDesirer.StopReturns(errors.New("failed-to-stop"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to stop app")))
			})
		})
	})

	Describe("Stop an app instance", func() {

		JustBeforeEach(func() {
			err = lrpBifrost.StopInstance(context.Background(), opi.LRPIdentifier{GUID: "guid_1234", Version: "version_1234"}, 1)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should call the desirer with the expected guid and index", func() {
			identifier, index := lrpDesirer.StopInstanceArgsForCall(0)
			Expect(identifier.GUID).To(Equal("guid_1234"))
			Expect(identifier.Version).To(Equal("version_1234"))
			Expect(index).To(Equal(uint(1)))
		})

		Context("when desirer's stop instance fails", func() {
			BeforeEach(func() {
				lrpDesirer.StopInstanceReturns(errors.New("failed-to-stop"))
			})

			It("returns a meaningful error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to stop instance")))
			})
		})
	})

	Describe("Get all instances of an app", func() {
		var (
			instances    []*cf.Instance
			opiInstances []*opi.Instance
		)

		BeforeEach(func() {
			opiInstances = []*opi.Instance{
				{Index: 0, Since: 123, State: opi.RunningState},
				{Index: 1, Since: 345, State: opi.CrashedState},
				{Index: 2, Since: 678, State: opi.ErrorState, PlacementError: "this is not the place"},
			}

			lrpDesirer.GetInstancesReturns(opiInstances, nil)
		})

		JustBeforeEach(func() {
			instances, err = lrpBifrost.GetInstances(context.Background(), opi.LRPIdentifier{GUID: "guid_1234", Version: "version_1234"})
		})

		It("should get the app instances from Desirer", func() {
			Expect(lrpDesirer.GetInstancesCallCount()).To(Equal(1))
			identifier := lrpDesirer.GetInstancesArgsForCall(0)
			Expect(identifier.GUID).To(Equal("guid_1234"))
			Expect(identifier.Version).To(Equal("version_1234"))
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return all running instances", func() {
			Expect(instances).To(Equal([]*cf.Instance{
				{Index: 0, Since: 123, State: opi.RunningState},
				{Index: 1, Since: 345, State: opi.CrashedState},
				{Index: 2, Since: 678, State: opi.ErrorState, PlacementError: "this is not the place"},
			}))
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				lrpDesirer.GetInstancesReturns([]*opi.Instance{}, errors.New("not found"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to get instances for app")))
			})
		})

	})

})
