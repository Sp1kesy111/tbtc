pragma solidity 0.5.17;

import {IBondedECDSAKeepFactory} from "@keep-network/keep-ecdsa/contracts/api/IBondedECDSAKeepFactory.sol";

/// @title Bonded ECDSA keep factory selection strategy.
/// @notice The strategy defines the algorithm for selecting a factory. tBTC
/// uses two bonded ECDSA keep factories, selecting one of them for each new
/// deposit being opened.
interface KeepFactorySelector {

    /// @notice Selects keep factory for the new deposit.
    /// @param _seed Request seed.
    /// @param _keepStakeFactory Regular, KEEP-stake based keep factory.
    /// @param _ethStakeFactory Fully backed, ETH-stake based keep factory.
    /// @return The selected keep factory.
    function selectFactory(
        uint256 _seed,
        IBondedECDSAKeepFactory _keepStakeFactory,
        IBondedECDSAKeepFactory _ethStakeFactory
    ) external view returns (IBondedECDSAKeepFactory);
}

/// @title Bonded ECDSA keep factory selection library.
/// @notice tBTC uses two bonded ECDSA keep factories: one based on KEEP stake
/// and ETH bond, and another based on ETH stake and ETH bond. The library holds
/// a reference to both factories as well as a reference to a selection strategy
/// deciding which factory to choose for the new deposit being opened.
library KeepFactorySelection {

    struct Storage {
        uint256 requestCounter;

        IBondedECDSAKeepFactory selectedFactory;

        KeepFactorySelector factorySelector;

        // Standard ECDSA keep factory: KEEP stake and ETH bond.
        IBondedECDSAKeepFactory keepStakeFactory;

        // Fully backed ECDSA keep factory: ETH stake and ETH bond.
        IBondedECDSAKeepFactory ethStakeFactory;
    }

    /// @notice Returns the selected keep factory.
    /// This function guarantees that the same factory is returned for every
    /// call until selectFactoryAndRefresh is executed. This lets to evaluate
    /// open keep fee estimate on the same factory that will be used later for
    /// opening a new keep (fee estimate and open keep requests are two
    /// separate calls).
    /// @return Selected keep factory. The same vale will be returned for every
    /// call of this function until selectFactoryAndRefresh is executed.
    function selectFactory(
        Storage storage _self
    ) public view returns (IBondedECDSAKeepFactory) {
        if (address(_self.selectedFactory) == address(0)) {
            return _self.keepStakeFactory; // TODO: we can use a setter
        }

        return _self.selectedFactory;
    }

    /// @notice Returns the selected keep factory and refreshes the choice
    /// for the next select call. The value returned by this function has been
    /// evaluated during the previous call. This lets to return the same value
    /// from selectFactory and selectFactoryAndRefresh, thus, allowing to use
    /// the same factory for which open keep fee estimate was evaluated (fee
    /// estimate and open keep requests are two separate calls).
    /// @return Selected keep factory.
    function selectFactoryAndRefresh(
        Storage storage _self
    ) public returns (IBondedECDSAKeepFactory) {
        IBondedECDSAKeepFactory factory = selectFactory(_self);
        refreshFactory(_self);

        return factory;
    }

    /// @notice Refreshes the keep factory choice. If either ETH-stake factory
    /// or selection strategy is not set, KEEP-stake factory is selected.
    /// Otherwise, calls selection strategy providing addresses of both
    /// factories to make a choice. Additionally, passes the selection seed
    /// evaluated from the current request counter value.
    function refreshFactory(Storage storage _self) internal {
        if (
            address(_self.ethStakeFactory) == address(0) ||
            address(_self.factorySelector) == address(0)
        ) {
            // KEEP-stake factory is guaranteed to be there. If the selection
            // can not be performed, this is the default choice.
            _self.selectedFactory = _self.keepStakeFactory;
            return;
        }

        _self.requestCounter++;
        uint256 seed = uint256(
            keccak256(abi.encodePacked(address(this), _self.requestCounter))
        );
        _self.selectedFactory = _self.factorySelector.selectFactory(
            seed,
            _self.keepStakeFactory,
            _self.ethStakeFactory
        );
    }

    /// @notice Sets the address of the fully backed ECDSA keep factory.
    /// @dev Beware, can be called only once!
    /// @param _fullyBackedFactory Address of the factory.
    function setFullyBackedKeepFactory(
        Storage storage _self,
        address _fullyBackedFactory
    ) internal {
        require(
            address(_self.ethStakeFactory) == address(0),
            "Fully backed factory address already set"
        );

        require(
            address(_fullyBackedFactory) != address(0),
            "Invalid address"
        );

        _self.ethStakeFactory = IBondedECDSAKeepFactory(_fullyBackedFactory);
    }

    /// @notice Sets the address of the keep factory selector contract.
    /// @dev Beware, can be called only once!
    /// @param _factorySelector Address of the keep factory selector contract.
    function setKeepFactorySelector(
         Storage storage _self,
        address _factorySelector
    ) internal {
        require(
            address(_self.factorySelector) == address(0),
            "Factory selector contract address already set"
        );

        require(
            address(_factorySelector) != address(0),
            "Invalid address"
        );

        _self.factorySelector = KeepFactorySelector(_factorySelector);
    }
}
