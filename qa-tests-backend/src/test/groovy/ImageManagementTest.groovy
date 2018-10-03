import groups.BAT
import groups.Integration
import org.junit.experimental.categories.Category
import spock.lang.Unroll
import stackrox.generated.PolicyServiceOuterClass.LifecycleStage

class ImageManagementTest extends BaseSpecification {
    @Unroll
    @Category([BAT, Integration])
    def "Verify CI/CD Integration Endpoint - #policy"() {
        when:
        "Update Policy to BUILD_TIME"
        def startStages = Services.updatePolicyLifecycleStage(policy, [LifecycleStage.BUILD_TIME,])

        and:
        "Request Image Scan"
        def scanResults = Services.requestBuildImageScan(imageRegistry, imageRemote, imageTag)

        then:
        "verify policy exists in response"
        assert scanResults.getAlertsList().findAll { it.getPolicy().name == policy }.size() == 1

        cleanup:
        "Revert Policy"
        Services.updatePolicyLifecycleStage(policy, startStages)

        where:
        "Data inputs are: "

        policy                                        | imageRegistry            | imageRemote              | imageTag
        "Latest tag"                                  | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        //intentionally use the same policy twice to make sure alert count does not increment
        "Latest tag"                                  | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "90-Day Image Age"                            | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "Aptitude Package Manager (apt) in Image"     | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "Curl in Image"                               | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "CVSS >= 7"                                   | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "Wget in Image"                               | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
        "Apache Struts: CVE-2017-5638"                | "apollo-dtr.rox.systems" | "legacy-apps/struts-app" | "latest"
    }

    @Category(Integration)
    def "Verify lifecycle Stage can only be BUILD_TIME for policies with image criteria"() {
        when:
        "Update Policy to BUILD_TIME"
        def startStages = Services.updatePolicyLifecycleStage(
                "No resource requests or limits specified",
                [LifecycleStage.BUILD_TIME,]
        )

        then:
        "assert startStage is null"
        assert startStages == []
    }
}
